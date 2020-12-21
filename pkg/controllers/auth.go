package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	"github.com/gorilla/mux"

	authapi "imanager/pkg/api/auth"
	"imanager/pkg/controllers/paser"
	authsvc "imanager/pkg/services/auth"
	"imanager/pkg/util"
)

type AuthController struct {
}

func (c AuthController) CreateTokenInHttp(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body read failed, %v", err))
		return
	}
	reqToken := authapi.ReqToken{}
	err = json.Unmarshal(requestBody, &reqToken)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body unmarshal failed, %v", err))
		return
	}

	glog.Infof("%v request token", reqToken.Auth.Name)
	isValid, user, err := authsvc.ValidUserPasswordAndGetRoles(reqToken.Auth.Name, reqToken.Auth.Password)
	if err != nil {
		glog.Errorf("valid user[%v]'s password failed, err: %v", reqToken.Auth.Name, err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("valid user's password failed, %v", err))
		return
	}
	if !isValid {
		glog.Errorf("user name[%v] or password is invalid", reqToken.Auth.Name)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("user name or password is invalid"))
		return
	}

	if reqToken.Scope.Duration == 0 {
		reqToken.Scope.Duration = authapi.DefaultExpireTime
	}
	reqToken.Scope.Duration = reqToken.Scope.Duration * authapi.BaseDuration

	issuedAt := time.Now()
	res := authapi.RespToken{
		ExpiresAt: issuedAt.Add(reqToken.Scope.Duration),
		IssuedAt:  issuedAt,
		Name:      reqToken.Auth.Name,
		UserID:    user.UUID,
		Role:      user.Role,
		Group:     user.Group,
		TrueName:  user.TruthName,
	}

	tokenss, err := authsvc.CreateToken(res)
	if err != nil {
		glog.Errorf("create token failed, user name: %v, err: %v", reqToken.Auth.Name, err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("create token failed, %v", err))
		return
	}

	respBody, _ := json.Marshal(res)
	w.Header().Set(authapi.TokenHeaderKey, tokenss)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBody)
}

func (c AuthController) CheckTokenInHttp(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get(authapi.TokenHeaderKey)
	info, err := authsvc.ParseToken(tokenStr)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("parse token failed, %v", err))
		return
	}
	respBody, _ := json.Marshal(info)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBody)
}

func isAllowedModifyUser(user *authapi.User, info *authapi.RespToken) bool {
	largestRolePermission := authapi.GetLargestRolePermission(info.Role)
	if largestRolePermission.IsLargerPermission(authapi.OpServiceRole) {
		return true
	}
	if !largestRolePermission.IsLargerPermission(authapi.AdminRole) {
		return false
	}

	group, err:= authsvc.GetGroupByID(info.Group.ID)
	if err != nil {
		glog.Errorf("get detail group by id failed, user: %v/%v, group id: %v, err: %v", user.Name, user.UUID, info.Group.ID, err)
		return false
	}
	for _, v := range group.User {
		if v.Name == user.Name {
			return true
		}
	}

	return false
}

var (
	UserNameRegexp      = "^[a-zA-Z0-9-]{4,64}$"
	PasswordRegexp      = "^[a-zA-Z0-9_-]{8,18}$"
	UserTruthNameRegexp = `^[a-zA-Z\p{Han}]+$`
	EmailRegexp         = `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	PhoneNumRegexp      = "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"
)

func validUserForCreateOrUpdate(user *authapi.User, isCreate bool, info *authapi.RespToken) error {
	var isMatch bool
	if isCreate || len(user.Name) != 0 {
		isMatch, _ = regexp.MatchString(UserNameRegexp, user.Name)
		if !isMatch {
			return fmt.Errorf("user name doesn't match the format")
		}
	}
	if !isCreate && len(user.Name) == 0 && len(user.UUID) == 0 {
		return fmt.Errorf("user name and uuid should not be empty at same time")
	}
	if isCreate || len(user.Password) != 0 {
		isMatch, _ = regexp.MatchString(PasswordRegexp, user.Password)
		if !isMatch {
			return fmt.Errorf("user password doesn't match the format")
		}
		hasLower := regexp.MustCompile(`[a-z]`)
		hasUpper := regexp.MustCompile(`[A-Z]`)
		hasNumber := regexp.MustCompile(`[0-9]`)
		if !(hasLower.MatchString(user.Password) && hasUpper.MatchString(user.Password) && hasNumber.MatchString(user.Password)) {
			return fmt.Errorf("user password must longer than 8 chars and shorter than 18 chars with at least 1 uppercase letter, 1 lowercase letter and 1 number")
		}
	}
	if isCreate || len(user.TruthName) != 0 {
		isMatch, _ = regexp.MatchString(UserTruthNameRegexp, user.TruthName)
		if !isMatch {
			return fmt.Errorf("user truth name doesn't match the format")
		}
	}
	if isCreate || len(user.Email) != 0 {
		isMatch, _ = regexp.MatchString(EmailRegexp, user.Email)
		if !isMatch {
			return fmt.Errorf("user email doesn't match the format")
		}
	}
	if isCreate || len(user.PhoneNum) != 0 {
		isMatch, _ = regexp.MatchString(PhoneNumRegexp, user.PhoneNum)
		if !isMatch {
			return fmt.Errorf("user phone num doesn't match the format")
		}
	}

	largestRolePermissionInInfo := authapi.GetLargestRolePermission(info.Role)
	if isCreate && user.Group == nil {
		if largestRolePermissionInInfo == authapi.OpServiceRole {
			user.Group = authsvc.DefaultGroup
		} else {
			user.Group = info.Group
		}
	}
	if isCreate && user.Role == nil {
		user.Role = authsvc.DefaultRole
	}
	if user.Role != nil {
		largestRolePermissionInUser := authapi.GetLargestRolePermission(user.Role)
		if largestRolePermissionInUser == authapi.OpServiceRole {
			user.Group = authsvc.OpServiceGroup
		}
		if !largestRolePermissionInInfo.IsLargerPermission(largestRolePermissionInUser) {
			return fmt.Errorf("user's permission is not allowed more authority than info")
		}
	}

	// op service can create user into any group and any role
	if largestRolePermissionInInfo == authapi.OpServiceRole {
		return nil
	}
	if user.Group != nil && user.Group.ID != info.Group.ID {
		return fmt.Errorf("user's group and info's group are not allowed to differ")
	}


	return nil
}

func (c AuthController) CreateUser(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}
	// only op service can create user
	if !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.AdminRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to create user")
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body read failed, %v", err))
		return
	}
	user := &authapi.User{}
	err = json.Unmarshal(requestBody, user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body unmarshal failed, %v", err))
		return
	}
	err = validUserForCreateOrUpdate(user, true, info)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}
	user, err = authsvc.CreateUser(user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("create user failed, %v", err))
		return
	}
	out, err := json.Marshal(user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal user failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(out)
}

func (c AuthController) ModifyUser(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body read failed, %v", err))
		return
	}
	user := &authapi.User{}
	err = json.Unmarshal(requestBody, user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body unmarshal failed, %v", err))
		return
	}
	glog.Infof("username: %v, info name: %v", user.Name, info.Name)
	if user.Name != info.Name && !isAllowedModifyUser(user, info) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to modify")
		return
	}
	err = validUserForCreateOrUpdate(user, false, info)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	err = authsvc.IsAllowUserUpdate(user, info)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	user, err = authsvc.UpdateUser(user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("update user failed, %v", err))
		return
	}
	out, err := json.Marshal(user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal user failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (c AuthController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}

	name := mux.Vars(r)["name"]
	// only op service and admin in user's group can delete user
	if !isAllowedModifyUser(&authapi.User{Name:name}, info) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to delete user")
		return
	}
	glog.Infof("delete user[%v] by %v/%v", name, info.Name, info.UserID)
	err = authsvc.DeleteUserByName(name)
	if err == orm.ErrNoRows {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "user isn't exist")
		return
	}
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("delete user failed, %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c AuthController) GetUser(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}

	name := mux.Vars(r)["name"]
	if name != info.Name && !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to get user detail")
		return
	}

	user, err := authsvc.GetUserByName(name)
	if err == orm.ErrNoRows {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "user isn't exist")
		return
	}
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("get user from db failed, %v", err))
		return
	}
	out, err := json.Marshal(user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal user failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (c AuthController) ListUser(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}
	dataSelect := paser.PaserDataSelectPathParameter(r)

	resp, num, err := authsvc.ListUserByUserID(getManageUserIDs(info), dataSelect)
	if err != nil {
		glog.Errorf("list users failed, query user: %v/%v, %v", info.Name, info.UserID, err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	respBody, _ := json.Marshal(authapi.UserList{
		Count: num,
		Item:  resp,
	})
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBody)
}

func (c AuthController) InitUser(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if len(name) == 0 {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, "user name is empty")
		return
	}
	_, err := authsvc.InitUser(name)
	if err == orm.ErrNoRows {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusNotFound, "user isn't exist")
		return
	}
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("encrypt user[%v] failed, %v", name, err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c AuthController) UnInitUser(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}
	// only op service can unInit user
	if !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to unInit user")
		return
	}

	name := mux.Vars(r)["name"]
	if len(name) == 0 {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, "user name is empty")
		return
	}
	_, err = authsvc.UnInitUser(name)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("encrypt user[%v] failed, %v", name, err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

var (
	NameRegexp       = "^[a-zA-Z0-9-]{1,64}$"
	AnnotationRegexp = `^[a-zA-Z0-9\p{Han}-_.,:/@#{}\\\"]+$`
	AnnotationMaxLen = 2048
)

func validRole(role *authapi.Role) error {
	isMatch, _ := regexp.MatchString(NameRegexp, role.Name)
	if !isMatch {
		return fmt.Errorf("role name don't match the format")
	}
	if len(role.Annotation) > AnnotationMaxLen {
		return fmt.Errorf("role annotation too long, max length is: %v", AnnotationMaxLen)
	}
	isMatch, _ = regexp.MatchString(AnnotationRegexp, role.Annotation)
	if !isMatch {
		return fmt.Errorf("role annotation don't match the format")
	}
	return nil
}

func (c AuthController) CreateRole(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}
	// only op service can create role
	if !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to create role")
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body read failed, %v", err))
		return
	}
	role := &authapi.Role{}
	err = json.Unmarshal(requestBody, role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body unmarshal failed, %v", err))
		return
	}

	err = validRole(role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	role, err = authsvc.CreateRole(role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("create role in db failed, %v", err))
		return
	}
	out, err := json.Marshal(role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(out)
}

func (c AuthController) ModifyRole(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}

	if !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to modify")
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body read failed, %v", err))
		return
	}
	role := &authapi.Role{}
	err = json.Unmarshal(requestBody, role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body unmarshal failed, %v", err))
		return
	}
	err = validRole(role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	role, err = authsvc.UpdateRole(role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("update role in db failed, %v", err))
		return
	}
	out, err := json.Marshal(role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal role failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (c AuthController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}
	// only op service can delete role
	if !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to delete role")
		return
	}
	name := mux.Vars(r)["name"]
	glog.Infof("delete role[%v] by %v/%v", name, info.Name, info.UserID)
	err = authsvc.DeleteRoleByName(name)
	if err == orm.ErrNoRows {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "role isn't exist")
		return
	}
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("delete role in db failed, %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c AuthController) GetRole(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	role, err := authsvc.GetRoleByName(name)
	if err == orm.ErrNoRows {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "role isn't exist")
		return
	}
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("get role from db failed, %v", err))
		return
	}
	out, err := json.Marshal(role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal role failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (c AuthController) ListRole(w http.ResponseWriter, r *http.Request) {
	dataSelect := paser.PaserDataSelectPathParameter(r)
	roles, num, err := authsvc.ListRole(dataSelect)
	if err != nil {
		glog.Errorf("list group failed, %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	respBody, _ := json.Marshal(authapi.RoleList{
		Count: num,
		Item:  roles,
	})
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBody)
}

func validGroup(group *authapi.Group) error {
	isMatch, _ := regexp.MatchString(NameRegexp, group.Name)
	if !isMatch {
		return fmt.Errorf("group name don't match the format")
	}
	if len(group.Annotation) > AnnotationMaxLen {
		return fmt.Errorf("group annotation too long, max length is: %v", AnnotationMaxLen)
	}
	isMatch, _ = regexp.MatchString(AnnotationRegexp, group.Annotation)
	if !isMatch {
		return fmt.Errorf("group annotation don't match the format")
	}
	return nil
}

func (c AuthController) CreateGroup(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}
	// only op service can create
	if !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to create group")
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body read failed, %v", err))
		return
	}
	group := &authapi.Group{}
	err = json.Unmarshal(requestBody, group)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body unmarshal failed, %v", err))
		return
	}

	err = validGroup(group)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	group, err = authsvc.CreateGroup(group)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("create group in db failed, %v", err))
		return
	}
	out, err := json.Marshal(group)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(out)
}

func (c AuthController) ModifyGroup(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}

	if !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to modify")
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body read failed, %v", err))
		return
	}
	group := &authapi.Group{}
	err = json.Unmarshal(requestBody, group)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("request body unmarshal failed, %v", err))
		return
	}
	err = validGroup(group)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	group, err = authsvc.UpdateGroup(group)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("update group in db failed, %v", err))
		return
	}
	out, err := json.Marshal(group)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (c AuthController) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}
	// only op service can delete group
	if !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to delete group")
		return
	}
	name := mux.Vars(r)["name"]
	glog.Infof("delete group[%v] by %v/%v", name, info.Name, info.UserID)
	err = authsvc.DeleteGroupByName(name)
	if err == orm.ErrNoRows {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "group isn't exist")
		return
	}
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("delete role in db failed, %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c AuthController) GetGroup(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}

	name := mux.Vars(r)["name"]
	if info.Group.Name != name && !authapi.GetLargestRolePermission(info.Role).IsLargerPermission(authapi.OpServiceRole) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to get group detail")
		return
	}

	role, err := authsvc.GetGroupByName(name)
	if err == orm.ErrNoRows {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "group isn't exist")
		return
	}
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("get group from db failed, %v", err))
		return
	}
	out, err := json.Marshal(role)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("marshal group failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (c AuthController) ListGroup(w http.ResponseWriter, r *http.Request) {
	dataSelect := paser.PaserDataSelectPathParameter(r)
	groups, num, err := authsvc.ListGroup(dataSelect)
	if err != nil {
		glog.Errorf("list group failed, %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	respBody, _ := json.Marshal(authapi.GroupList{
		Count: num,
		Item:  groups,
	})
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBody)
}
