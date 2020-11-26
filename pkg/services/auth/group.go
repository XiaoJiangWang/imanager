package auth

import (
	"github.com/astaxie/beego/orm"
	"github.com/golang/glog"

	authapi "imanager/pkg/api/auth"
	"imanager/pkg/api/dataselect"
	"imanager/pkg/db/auth"
	authdb "imanager/pkg/db/auth"
)

func GetGroupByID(id int) (*auth.Group, error) {
	o := orm.NewOrm()
	group, err := auth.GetGroupByID(o, id)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func ListGroup(query *dataselect.DataSelectQuery) ([]authapi.Group, int64, error) {
	groupInDBs, nums, err := authdb.ListGroup(orm.NewOrm(), query)
	if err == orm.ErrNoRows {
		glog.Errorf("can't list group in db, no rows in db")
		return []authapi.Group{}, 0, nil
	}
	if err != nil {
		glog.Errorf("can't list group in db, err: %v", err)
		return []authapi.Group{}, 0, err
	}
	res := transformGroupDBs2APIs(groupInDBs)
	return res, nums, nil
}

func GetGroupByName(name string) (authapi.Group, error) {
	group, err := authdb.GetGroupByName(orm.NewOrm(), name)
	if err != nil {
		glog.Errorf("get role failed, err: %v", err)
		return authapi.Group{}, err
	}
	res := transformGroupDB2API(group)
	return res, nil
}