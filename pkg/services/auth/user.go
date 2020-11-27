package auth

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"
	uuid "github.com/satori/go.uuid"

	authapi "imanager/pkg/api/auth"
	"imanager/pkg/api/dataselect"
	authdb "imanager/pkg/db/auth"
	"imanager/pkg/encrypt"
)

var (
	DefaultGroup = &authapi.GroupInUser{ID: 1}
	DefaultRole  = []authapi.RoleInUser{{ID: 3}}
)

func ValidUserPasswordAndGetRoles(name, password string) (bool, *authdb.User, error) {
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

	return true, &user, nil
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
	user.Password, err = encrypt.Decrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		glog.Errorf("decrypt user password failed, user: %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	userApi := transformUserDB2API(user)
	return &userApi, nil
}

func UpdateUser(user *authapi.User) (*authapi.User, error) {
	var err error
	user.Password, err = encrypt.Encrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		glog.Errorf("encrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}

	userDB := transformUserAPI2DB(*user)
	o := orm.NewOrm()
	userDB, err = authdb.UpdateUser(o, userDB)
	if err != nil {
		glog.Errorf("update user[%v/%v] failed, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	newUser := transformUserDB2API(userDB)
	newUser.Password, err = encrypt.Decrypt(newUser.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		glog.Errorf("decrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
	return &newUser, nil
}

func DeleteUserByName(name string) error {
	return authdb.DeleteUserByName(orm.NewOrm(), name)
}

func CreateUser(user *authapi.User) (*authapi.User, error) {
	var err error
	user.UUID = uuid.NewV4().String()
	user.Password, err = encrypt.Encrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
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

	userDB := transformUserAPI2DB(*user)
	userDB, err = authdb.CreateUser(orm.NewOrm(), userDB)
	if err != nil {
		glog.Errorf("create user[%v] failed, err: %v", user.Name, err)
		return nil, err
	}
	newUser := transformUserDB2API(userDB)
	user.Password, err = encrypt.Decrypt(user.Password, encrypt.CpabeType, encrypt.OpServiceRole)
	if err != nil {
		glog.Errorf("decrypt password failed for %v/%v, err: %v", user.Name, user.UUID, err)
		return nil, err
	}
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
	res := transformUserDBs2APIs(userInDBs)
	return res, nums, nil
}
