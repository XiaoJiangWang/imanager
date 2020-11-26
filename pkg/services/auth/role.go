package auth

import (
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