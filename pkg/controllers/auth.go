package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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
		Roles:     user.Role,
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

func hasSuperPermission(info *authapi.RespToken) bool {
	for _, v := range info.Roles {
		if authapi.RoleType(v.Name) == authapi.OpServiceRole {
			return true
		}
	}

	return false
}

func validUserForCreateOrUpdate(info *authapi.RespToken, user *authapi.User) error {
	var userUpPermission, infoUpPermission = authapi.UserRole, authapi.UserRole
	for _, v := range user.Role {
		if !userUpPermission.IsLargerPermission(authapi.RoleType(v.Name)) {
			userUpPermission = authapi.RoleType(v.Name)
		}
	}
	for _, v := range info.Roles {
		if !infoUpPermission.IsLargerPermission(authapi.RoleType(v.Name)) {
			infoUpPermission = authapi.RoleType(v.Name)
		}
	}
	if userUpPermission.IsLargerPermission(infoUpPermission) {
		return errors.New("permission denied")
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
	if !hasSuperPermission(info) {
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
	user, err = authsvc.CreateUser(user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("create user in db failed, %v", err))
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

	name := mux.Vars(r)["name"]
	if name != info.Name && !hasSuperPermission(info) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to modify")
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

	user, err = authsvc.UpdateUser(user)
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("update user in db failed, %v", err))
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
	// only op service can delete user
	if !hasSuperPermission(info) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to delete user detail")
		return
	}
	name := mux.Vars(r)["name"]
	glog.Infof("delete user[%v] by %v/%v", name, info.Name, info.UserID)
	err = authsvc.DeleteUserByName(name)
	if err == orm.ErrNoRows {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "user isn't exist")
		return
	}
	if err != nil {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusInternalServerError, fmt.Sprintf("delete user in db failed, %v", err))
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
	if name != info.Name && !hasSuperPermission(info) {
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

func (c AuthController) CreateRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("TODO"))
}

func (c AuthController) ModifyRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("TODO"))
}

func (c AuthController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("TODO"))
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

func (c AuthController) CreateGroup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("TODO"))
}

func (c AuthController) ModifyGroup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("TODO"))
}

func (c AuthController) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("TODO"))
}

func (c AuthController) GetGroup(w http.ResponseWriter, r *http.Request) {
	info, err := util.GetUserInfo(r.Header.Get(authapi.ParseInfo))
	if err != nil {
		glog.Errorf("get user info from header failed, err: %v", err)
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, err.Error())
		return
	}

	if !hasSuperPermission(info) {
		util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, "no permission to get group detail")
		return
	}

	name := mux.Vars(r)["name"]
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
