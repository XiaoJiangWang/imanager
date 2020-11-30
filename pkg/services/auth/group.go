package auth

import (
	"fmt"

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

func CreateGroup(group *authapi.Group) (*authapi.Group, error) {
	var err error
	groupInDB := transformGroupAPI2DB(*group)
	groupInDB, err = authdb.CreateGroup(orm.NewOrm(), groupInDB)
	if err != nil {
		glog.Errorf("create group[%v] failed, err: %v", group.Name, err)
		return nil, err
	}
	newGroup := transformGroupDB2API(groupInDB)
	return &newGroup, nil
}

func UpdateGroup(group *authapi.Group) (*authapi.Group, error) {
	var err error
	groupInDB := transformGroupAPI2DB(*group)
	groupInDB, err = authdb.UpdateGroup(orm.NewOrm(), groupInDB)
	if err != nil {
		glog.Errorf("update group[%v/%v] failed, err: %v", group.Name, group.ID, err)
		return nil, err
	}
	newGroup := transformGroupDB2API(groupInDB)
	return &newGroup, nil
}

func DeleteGroupByName(name string) error {
	o := orm.NewOrm()
	group, err := authdb.GetGroupByName(o, name)
	if err != nil {
		return err
	}
	if len(group.User) != 0 {
		return fmt.Errorf("group is in use, can't delete")
	}
	return authdb.DeleteGroupByName(o, name)
}
