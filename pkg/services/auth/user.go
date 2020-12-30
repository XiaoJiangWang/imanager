package auth

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	uuid "github.com/satori/go.uuid"

	authapi "imanager/pkg/api/auth"
	"imanager/pkg/api/dataselect"
	authdb "imanager/pkg/db/auth"
	"imanager/pkg/encrypt"
	"imanager/pkg/services/util"
)

var (
	DefaultGroup   = &authapi.GroupInUser{ID: 1}
	OpServiceGroup = &authapi.GroupInUser{ID: 999}
	DefaultRole    = []authapi.RoleInUser{{ID: 3}}
)

func ValidUserPasswordAndGetRoles(name, password string) (bool, *authapi.User, error) {
	o := orm.NewOrm()
	user, err := authdb.GetUserByName(o, name)
	if err == orm.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	user.Password, err = encrypt.Decrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		return false, nil, fmt.Errorf("decrypt failed, %v", err)
	}
	if user.Password != password {
		return false, nil, nil
	}
	out := transformUserDB2API(user)

	return true, &out, nil
}

func IsAllowUserUpdate(user *authapi.User, info *authapi.RespToken) error {
	var err error
	o := orm.NewOrm()

	// 用户角色校验
	var userUpPermission, infoUpPermission = authapi.UserRole, authapi.UserRole
	for _, v := range user.Role {
		tmp, err := authdb.GetRoleByID(o, v.ID)
		if err != nil {
			return fmt.Errorf("get role from db failed, %v", err)
		}
		if !userUpPermission.IsLargerPermission(authapi.RoleType(tmp.Id)) {
			userUpPermission = authapi.RoleType(tmp.Id)
		}
	}
	for _, v := range info.Role {
		if !infoUpPermission.IsLargerPermission(authapi.RoleType(v.ID)) {
			infoUpPermission = authapi.RoleType(v.ID)
		}
	}
	if !infoUpPermission.IsLargerPermission(userUpPermission) {
		return errors.New("no permission to modify role")
	}

	// 组校验
	if infoUpPermission == authapi.OpServiceRole || user.Group == nil {
		return nil
	}

	var oldUser authdb.User
	if len(user.Name) != 0 {
		oldUser, err = authdb.GetUserByName(o, user.Name)
	} else if len(user.UUID) != 0 {
		oldUser, err = authdb.GetUserByUUID(o, user.UUID)
	}
	if err != nil {
		return fmt.Errorf("get user detail failed, %v", err)
	}
	if user.Group.ID != oldUser.Group.Id {
		return errors.New("no permission to modify group")
	}
	return nil
}

func GetUserByUUID(uuid string) (*authdb.User, error) {
	o := orm.NewOrm()
	user, err := authdb.GetUserByUUID(o, uuid)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByName(name string) (*authapi.User, error) {
	o := orm.NewOrm()
	user, err := authdb.GetUserByName(o, name)
	if err != nil {
		glog.Errorf("get user from db failed, name: %v, err: %v", name, err)
		return nil, err
	}
	//user.Password = ""
	user.Password, err = encrypt.Decrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		glog.Errorf("decrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	userApi := transformUserDB2API(user)
	return &userApi, nil
}

func UpdateUser(user *authapi.User) (*authapi.User, error) {
	var err error
	newPassword := user.Password
	if len(user.Password) != 0 {
		user.Password, err = encrypt.Encrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
		if err != nil {
			glog.Errorf("encrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
			return nil, err
		}
	}

	userDB := transformUserAPI2DB(*user)
	o := orm.NewOrm()
	err = o.Begin()
	if err != nil {
		return user, err
	}

	var oldUser authdb.User
	if userDB.UUID != "" {
		oldUser, err = authdb.GetUserByUUID(o, userDB.UUID)
	} else if userDB.Name != "" {
		oldUser, err = authdb.GetUserByName(o, userDB.Name)
	} else {
		_ = o.Commit()
		glog.Errorf("find user by name or uuid failed")
		return user, fmt.Errorf("find user by name or uuid failed")
	}
	if err != nil {
		_ = o.Rollback()
		return user, err
	}

	err = util.Patch(&oldUser, &userDB)
	if err != nil {
		_ = o.Rollback()
		return user, err
	}
	if len(userDB.Role) == 0 {
		userDB.Role = oldUser.Role
	}

	userDB, err = authdb.UpdateUser(o, userDB)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("update user[%v/%v] failed, err: %v", user.Name, user.UUID, err)
		return nil, err
	}

	// update in harbor
	userForHarbor := transformUserDB2API(userDB)
	oldUserForHarbor := transformUserDB2API(oldUser)
	if len(newPassword) != 0 {
		userForHarbor.Password = newPassword
		oldUserForHarbor.Password, err = encrypt.Decrypt(oldUserForHarbor.Password, encrypt.CpabeType, encrypt.OpServiceRole)
		if err != nil {
			glog.Errorf("encrypt old user password for harbor failed for %v/%v, err: %v", user.Name, user.UUID, err)
			return nil, err
		}
	}
	err = updateUserInHarbor(&oldUserForHarbor, &userForHarbor)
	if err != nil {
		glog.Errorf("update user in harbor failed, err: %v", err)
		_ = o.Rollback()
		return user, err
	}

	_ = o.Commit()

	//userDB.Password = ""
	userDB.Password, err = encrypt.Decrypt(userDB.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		glog.Errorf("decrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	newUser := transformUserDB2API(userDB)
	return &newUser, nil
}

func DeleteUserByName(name string) error {
	var err error
	o := orm.NewOrm()
	err = o.Begin()
	if err != nil {
		return err
	}

	err = authdb.DeleteUserByName(o, name)
	if err != nil {
		glog.Errorf("delete user in db failed, user: %v, err: %v", name, err)
		_ = o.Rollback()
		return err
	}

	err = deleteUserInHarbor(name)
	if err != nil {
		glog.Errorf("delete user in harbor failed, user: %v, err: %v", name, err)
		_ = o.Rollback()
		return err
	}

	_ = o.Commit()
	return nil
}

func CreateUser(user *authapi.User) (*authapi.User, error) {
	var err error
	user.UUID = uuid.NewV4().String()
	encryptPassword, err := encrypt.Encrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		glog.Errorf("encrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}

	if user.Group == nil {
		user.Group = DefaultGroup
	}

	if user.Role == nil || len(user.Role) == 0 {
		user.Role = DefaultRole
	}

	o := orm.NewOrm()
	err = o.Begin()
	if err != nil {
		return user, err
	}


	userDB := transformUserAPI2DB(*user)
	userDB.Password = encryptPassword
	userDB, err = authdb.CreateUser(o, userDB)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("create user[%v] failed, err: %v", user.Name, err)
		return nil, err
	}

	err = createUserInHarbor(user)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("create user[%v] in harbor failed, err: %v", user.Name, err)
		return nil, err
	}

	_ = o.Commit()

	//userDB.Password = ""
	userDB.Password, err = encrypt.Decrypt(userDB.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		glog.Errorf("decrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	newUser := transformUserDB2API(userDB)
	return &newUser, nil
}

func ListUserByUserID(userIDs []string, query *dataselect.DataSelectQuery) ([]authapi.User, int64, error) {
	userInDBs, nums, err := authdb.ListUsersByUserIDs(orm.NewOrm(), userIDs, query)
	if err == orm.ErrNoRows {
		glog.Errorf("can't list user[%v] in db, no rows in db", strings.Join(userIDs, ","))
		return []authapi.User{}, 0, nil
	}
	if err != nil {
		glog.Errorf("can't list user[%v] in db, err: %v", strings.Join(userIDs, ","), err)
		return []authapi.User{}, 0, err
	}
	for k := range userInDBs {
		//userInDBs[k].Password = ""
		userInDBs[k].Password, err = encrypt.Decrypt(userInDBs[k].Password, encrypt.CpabeType, encrypt.OpServiceRole)
		if err != nil {
			glog.Errorf("decrypt password failed for %v/%v, err: %v", userInDBs[k].Name, userInDBs[k].UUID, err)
			continue
		}
	}
	res := transformUserDBs2APIs(userInDBs)
	return res, nums, nil
}

var PasswordRegexp  = "^[a-zA-Z0-9_-]{8,18}$"

func InitUser(name string) (*authapi.User, error) {
	var err error
	o := orm.NewOrm()
	err = o.Begin()
	if err != nil {
		return nil, err
	}
	user, err := authdb.GetUserByName(o, name)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("get user from db failed, name: %v, err: %v", name, err)
		return nil, err
	}
	isMatch, _ := regexp.MatchString(PasswordRegexp, user.Password)
	if !isMatch {
		_ = o.Rollback()
		return nil, fmt.Errorf("user password doesn't match the format, maybe it already was encrypted, please check it in db")
	}
	user.Password, err = encrypt.Encrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("encrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	if len(user.UUID) == 0 {
		user.UUID = uuid.NewV4().String()
	}
	user, err = authdb.UpdateUser(o, user)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("update user[%v/%v] failed, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	userApi := transformUserDB2API(user)
	_ = o.Commit()
	return &userApi, nil
}

func UnInitUser(name string) (*authapi.User, error) {
	var err error
	o := orm.NewOrm()
	err = o.Begin()
	if err != nil {
		return nil, err
	}
	user, err := authdb.GetUserByName(o, name)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("get user from db failed, name: %v, err: %v", name, err)
		return nil, err
	}
	user.Password, err = encrypt.Decrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("encrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	user, err = authdb.UpdateUser(o, user)
	if err != nil {
		_ = o.Rollback()
		glog.Errorf("update user[%v/%v] failed, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	userApi := transformUserDB2API(user)
	_ = o.Commit()
	return &userApi, nil
}