package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"

	authapi "imanager/pkg/api/auth"
	"imanager/pkg/config"
)

var (
	client         *http.Client
	harborUsername string
	harborPassword string
	harborAddress  string
)

func init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	client = &http.Client{Transport: tr}
	harborUsername, harborPassword, harborAddress = config.GetConfig().String(usernameKey), config.GetConfig().String(passwordKey), config.GetConfig().String(harborAddressKey)
}

const (
	tokenURL              = "%s/service/token?account=admin&service=harbor-registry&scope=registry:catalog:*"
	createUserURL         = "%s/api/v2.0/users"
	updateUserProfileURL  = "%s/api/v2.0/users/%d"
	updateUserPasswordURL = "%s/api/v2.0/users/%d/password"
	updateUserSysAdminURL = "%s/api/v2.0/users/%d/sysadmin"
	deleteUserURL         = "%s/api/v2.0/users/%d"
	searchUserURL         = "%s/api/v2.0/users/search?username=%s"

	tokenKey         = "X-Harbor-CSRF-Token"
	usernameKey      = "HarborUser"
	passwordKey      = "HarborPassword"
	harborAddressKey = "HarborAddress"
)

func getTokenInHarbor() (string, error) {
	// curl -ikL -X GET -u wangxj:Harbor123456 http://10.5.26.86:8080/service/token?account=admin\&service=harbor-registry\&scope=registry:catalog:* -v
	url := fmt.Sprintf(tokenURL, harborAddress)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		glog.Errorf("create token request failed, err: %v", err)
		return "", err
	}

	req.SetBasicAuth(harborUsername, harborPassword)
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("create token request failed, err: %v", err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		glog.Errorf("create token request failed, err: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("read resp body failed, err: %v", err)
		return "", err
	}
	var token authapi.TokenInHarbor
	err = json.Unmarshal(body, &token)
	if err != nil {
		glog.Errorf("read resp body failed, err: %v", err)
		return "", err
	}
	return token.Token, nil
}

func createUserInHarbor(user *authapi.User) error {
	token, err := getTokenInHarbor()
	if err != nil {
		glog.Errorf("err: %v", err)
		return err
	}

	largestRole := authapi.GetLargestRolePermission(user.Role)
	nowTime := time.Now()
	harborUser := authapi.UserInHarbor{
		Username:        user.Name,
		Password:        user.Password,
		RealName:        user.TruthName,
		Email:           user.Email,
		AdminRoleInAuth: true,
		CreationTime:    nowTime,
		UpdateTime:      nowTime,
		SysadminFlag:    largestRole == authapi.OpServiceRole,
	}
	reqBody, _ := json.Marshal(harborUser)

	url := fmt.Sprintf(createUserURL, harborAddress)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		glog.Errorf("err: %v", err)
		return err
	}
	req.Header.Set(tokenKey, token)
	req.SetBasicAuth(harborUsername, harborPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("err: %v", err)
		return err
	}

	if resp.StatusCode == http.StatusCreated {
		glog.Infof("create success")
		return nil
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	var errorMessage authapi.ErrorMessageInHarbor
	err = json.Unmarshal(respBody, &errorMessage)
	if err != nil || len(errorMessage.Errors) == 0 {
		return fmt.Errorf("%v", string(respBody))
	}

	return fmt.Errorf("%v", errorMessage.Errors[0].Message)
}

func updateUserInHarbor(oldUser, user *authapi.User) error {
	// get token
	token, err := getTokenInHarbor()
	if err != nil {
		glog.Errorf("get token in harbor failed, err: %v", err)
		return err
	}

	needUpdateUserPassword := isNeedUpdateUserPassword(oldUser, user)
	needUpdateUserProfile := isNeedUpdateUserProfile(oldUser, user)
	needUpdateUserSysAdmin := isNeedUpdateUserSysAdmin(oldUser, user)
	if !(needUpdateUserPassword || needUpdateUserProfile || needUpdateUserSysAdmin) {
		glog.Infof("user[%v/%v] doesn't need update in harbor", user.Name, user.UUID)
		return nil
	}

	// get user id
	userID, err := searchUserInHarborByUserName(user.Name, token)
	if err != nil {
		glog.Errorf("get user id in harbor failed, user: [%v/%v], err: %v", user.Name, user.UUID, err)
		return err
	}

	// update user profile
	if needUpdateUserProfile {
		err = updateUserProfileInHarbor(userID, token, user)
		if err != nil {
			glog.Errorf("update user profile in harbor failed, user: [%v/%v], err: %v", user.Name, user.UUID, err)
			return err
		}
	}

	// update user password
	if needUpdateUserPassword {
		err = updateUserPasswordInHarbor(userID, token, user)
		if err != nil {
			glog.Errorf("update user password in harbor failed, user: [%v/%v], err: %v", user.Name, user.UUID, err)
			return err
		}
	}

	// update user to sysadmin
	if needUpdateUserSysAdmin {
		err = updateUser2SysAdminInHarbor(userID, token, user)
		if err != nil {
			glog.Errorf("update user sys admin in harbor failed, user: [%v/%v], err: %v", user.Name, user.UUID, err)
			return err
		}
	}

	return nil
}

func searchUserInHarborByUserName(username string, token string) (int, error) {
	URL := fmt.Sprintf(searchUserURL, harborAddress, username)
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set(tokenKey, token)
	req.SetBasicAuth(harborUsername, harborPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("err: %v", err)
		return 0, err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var tmp authapi.RespForSearchUserIDInHarbor
		err = json.Unmarshal(respBody, &tmp)
		if err != nil {
			return 0, err
		}
		if len(tmp) == 0 {
			return 0, fmt.Errorf("the user[%v] isn't exist in harbor", username)
		}
		for _, userInHarbor := range tmp {
			if userInHarbor.Username == username {
				glog.Infof("user[%v] id in harbor is %v", username, userInHarbor.UserID)
				return userInHarbor.UserID, nil
			}
			glog.Errorf("not found user in harbor, the num is %v, username: %v", len(tmp), username)
			return 0, fmt.Errorf("the user[%v] isn't exist in harbor", username)
		}
	}

	var errorMessage authapi.ErrorMessageInHarbor
	err = json.Unmarshal(respBody, &errorMessage)
	if err != nil || len(errorMessage.Errors) == 0 {
		return 0, fmt.Errorf("%v", string(respBody))
	}

	return 0, fmt.Errorf("%v", errorMessage.Errors[0].Message)
}

func updateUserProfileInHarbor(userID int, token string, user *authapi.User) error {
	harborUser := authapi.UserUpdateProfileInHarbor{
		Email:    user.Email,
		RealName: user.TruthName,
	}
	reqBody, _ := json.Marshal(harborUser)

	URL := fmt.Sprintf(updateUserProfileURL, harborAddress, userID)
	req, err := http.NewRequest(http.MethodPut, URL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set(tokenKey, token)
	req.SetBasicAuth(harborUsername, harborPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		glog.Infof("update user[%v/%v] profile success", user.Name, user.UUID)
		return nil
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	var errorMessage authapi.ErrorMessageInHarbor
	err = json.Unmarshal(respBody, &errorMessage)
	if err != nil || len(errorMessage.Errors) == 0 {
		return fmt.Errorf("%v", string(respBody))
	}

	return fmt.Errorf("%v", errorMessage.Errors[0].Message)
}

func updateUserPasswordInHarbor(userID int, token string, user *authapi.User) error {
	harborUser := authapi.UserUpdatePasswordInHarbor{
		NewPassword: user.Password,
	}
	reqBody, _ := json.Marshal(harborUser)

	URL := fmt.Sprintf(updateUserPasswordURL, harborAddress, userID)
	req, err := http.NewRequest(http.MethodPut, URL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set(tokenKey, token)
	req.SetBasicAuth(harborUsername, harborPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		glog.Infof("update user[%v/%v] profile success", user.Name, user.UUID)
		return nil
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	var errorMessage authapi.ErrorMessageInHarbor
	err = json.Unmarshal(respBody, &errorMessage)
	if err != nil || len(errorMessage.Errors) == 0 {
		return fmt.Errorf("%v", string(respBody))
	}

	return fmt.Errorf("%v", errorMessage.Errors[0].Message)
}

func updateUser2SysAdminInHarbor(userID int, token string, user *authapi.User) error {
	harborUser := authapi.UserUpdateSysAdminInHarbor{SysadminFlag: authapi.GetLargestRolePermission(user.Role) == authapi.OpServiceRole}
	reqBody, _ := json.Marshal(harborUser)

	URL := fmt.Sprintf(updateUserSysAdminURL, harborAddress, userID)
	req, err := http.NewRequest(http.MethodPut, URL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set(tokenKey, token)
	req.SetBasicAuth(harborUsername, harborPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		glog.Infof("update user[%v/%v] profile success", user.Name, user.UUID)
		return nil
	}

	var errorMessage authapi.ErrorMessageInHarbor
	err = json.Unmarshal(respBody, &errorMessage)
	if err != nil || len(errorMessage.Errors) == 0 {
		return fmt.Errorf("%v", string(respBody))
	}

	return fmt.Errorf("%v", errorMessage.Errors[0].Message)
}

func isNeedUpdateUserProfile(oldUser, user *authapi.User) bool {
	if oldUser.TruthName != user.TruthName {
		return true
	}
	if oldUser.Email != user.Email {
		return true
	}
	return false
}

func isNeedUpdateUserPassword(oldUser, user *authapi.User) bool {
	return oldUser.Password != user.Password
}

func isNeedUpdateUserSysAdmin(oldUser, user *authapi.User) bool {
	oldLargestRole := authapi.GetLargestRolePermission(oldUser.Role)
	newLargestRole := authapi.GetLargestRolePermission(user.Role)
	if oldLargestRole == newLargestRole {
		return false
	}
	if oldLargestRole == authapi.OpServiceRole || newLargestRole == authapi.OpServiceRole {
		return true
	}
	return false
}

func deleteUserInHarbor(username string) error {
	// get token
	token, err := getTokenInHarbor()
	if err != nil {
		glog.Errorf("get token in harbor failed, err: %v", err)
		return err
	}

	// get user id
	userID, err := searchUserInHarborByUserName(username, token)
	if err != nil {
		glog.Errorf("get user id in harbor failed, user: [%v], err: %v", username, err)
		return err
	}

	// delete user
	URL := fmt.Sprintf(deleteUserURL, harborAddress, userID)
	req, err := http.NewRequest(http.MethodDelete, URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set(tokenKey, token)
	req.SetBasicAuth(harborUsername, harborPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		glog.Infof("delete user[%v] profile success", username)
		return nil
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	var errorMessage authapi.ErrorMessageInHarbor
	err = json.Unmarshal(respBody, &errorMessage)
	if err != nil || len(errorMessage.Errors) == 0 {
		return fmt.Errorf("%v", string(respBody))
	}

	return fmt.Errorf("%v", errorMessage.Errors[0].Message)
}
