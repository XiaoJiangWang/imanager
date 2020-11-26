package controllers

import (
	"github.com/golang/glog"

	authapi "imanager/pkg/api/auth"
	"imanager/pkg/services/auth"
)

func getManageUserIDs(info *authapi.RespToken) []string {
	user, err := auth.GetUserByUUID(info.UserID)
	if err != nil {
		glog.Errorf("get detail user by uuid failed, user: %v/%v, err: %v", info.Name, info.UserID, err)
		return []string{info.UserID}
	}

	baseRole := authapi.UserRole
	for _, v := range user.Role {
		role := authapi.RoleType(v.Name)
		if role == authapi.OpServiceRole {
			// return all batch work
			return []string{}
		}
		if role.IsLargerPermission(baseRole) {
			baseRole = role
		}

	}
	if baseRole == authapi.UserRole {
		return []string{info.UserID}
	}

	group, err:= auth.GetGroupByID(user.Group.Id)
	if err != nil {
		glog.Errorf("get detail group by id failed, user: %v/%v, group id: %v, err: %v", info.Name, info.UserID, user.Group.Id, err)
		return []string{info.UserID}
	}

	res := make([]string, 0, len(group.User))
	for _, v := range group.User {
		res = append(res, v.UUID)
	}

	return res
}