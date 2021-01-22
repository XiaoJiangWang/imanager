package controllers

import (
	"github.com/golang/glog"

	authapi "imanager/pkg/api/auth"
	authsvc "imanager/pkg/services/auth"
)

func getManageUserIDs(info *authapi.RespToken) []string {
	user, err := authsvc.GetUserByUUID(info.UserID)
	if err != nil {
		glog.Errorf("get detail user by uuid failed, user: %v/%v, err: %v", info.Name, info.UserID, err)
		return []string{info.UserID}
	}

	baseRole := authapi.UserRole
	for _, v := range user.Role {
		role := authapi.RoleType(v.ID)
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

	group, err:= authsvc.GetGroupByID(user.Group.ID)
	if err != nil {
		glog.Errorf("get detail group by id failed, user: %v/%v, group id: %v, err: %v", info.Name, info.UserID, user.Group.ID, err)
		return []string{info.UserID}
	}

	res := make([]string, 0, len(group.User))
	for _, v := range group.User {
		res = append(res, v.UUID)
	}

	return res
}