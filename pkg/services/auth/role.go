package auth

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"

	authapi "imanager/pkg/api/auth"
	"imanager/pkg/api/dataselect"
	authdb "imanager/pkg/db/auth"
)

func ListRole(query *dataselect.DataSelectQuery) ([]authapi.Role, int64, error) {
	roleInDBs, nums, err := authdb.ListRole(orm.NewOrm(), query)
	if err == orm.ErrNoRows {
		glog.Errorf("can't list group in db, no rows in db")
		return []authapi.Role{}, 0, nil
	}
	if err != nil {
		glog.Errorf("can't list group in db, err: %v", err)
		return []authapi.Role{}, 0, err
	}
	res := transformRoleDBs2APIs(roleInDBs)
	return res, nums, nil
}

func GetRoleByName(name string) (authapi.Role, error) {
	role, err := authdb.GetRoleByName(orm.NewOrm(), name)
	if err != nil {
		glog.Errorf("get role failed, err: %v", err)
		return authapi.Role{}, err
	}
	res := transformRoleDB2API(role)
	return res, nil
}

func CreateRole(role *authapi.Role) (*authapi.Role, error) {
	var err error
	roleInDB := transformRoleAPI2DB(*role)
	roleInDB, err = authdb.CreateRole(orm.NewOrm(), roleInDB)
	if err != nil {
		glog.Errorf("create role[%v] failed, err: %v", role.Name, err)
		return nil, err
	}
	newRole := transformRoleDB2API(roleInDB)
	return &newRole, nil
}

func UpdateRole(role *authapi.Role) (*authapi.Role, error) {
	var err error
	roleInDB := transformRoleAPI2DB(*role)
	roleInDB, err = authdb.UpdateRole(orm.NewOrm(), roleInDB)
	if err != nil {
		glog.Errorf("update role[%v/%v] failed, err: %v", role.Name, role.ID, err)
		return nil, err
	}
	newRole := transformRoleDB2API(roleInDB)
	return &newRole, nil
}

func DeleteRoleByName(name string) error {
	o := orm.NewOrm()
	role, err := authdb.GetRoleByName(o, name)
	if err != nil {
		return err
	}
	if len(role.Group) != 0 || len(role.User) != 0 {
		return fmt.Errorf("role is in use, can't delete")
	}
	return authdb.DeleteRoleByName(o, name)
}