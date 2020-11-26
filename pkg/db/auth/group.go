package auth

import (
	"github.com/astaxie/beego/orm"

	"imanager/pkg/api/dataselect"
	"imanager/pkg/db/util"
)

type Group struct {
	Id         int     `json:"id" orm:"unique"`
	Name       string  `json:"name" orm:"unique"`
	Annotation string  `json:"annotation"`
	Role       []*Role `json:"role" orm:"rel(m2m)"`
	User       []*User `orm:"reverse(many)"`
}

func GetGroupByName(o orm.Ormer, name string) (Group, error) {
	group := Group{}
	err := o.QueryTable(Group{}).Filter("name", name).One(&group)
	return group, err
}

func GetGroupByID(o orm.Ormer, id int) (Group, error) {
	group := Group{Id: id}
	err := o.Read(&group)
	if err != nil {
		return group, err
	}
	_, err = o.LoadRelated(&group, "role")
	if err != nil {
		return group, err
	}
	_, err = o.LoadRelated(&group, "user")
	if err != nil {
		return group, err
	}
	return group, err
}

func UpdateGroup(o orm.Ormer, group Group) error {
	_, err := o.Update(&group)
	return err
}

func DeleteGroupByName(o orm.Ormer, name string) error {
	group := Group{Name: name}
	_, err := o.Delete(&group)
	return err
}

func CreateGroup(o orm.Ormer, group Group) (Group, error) {
	_, err := o.Insert(group)
	if err != nil {
		return group, err
	}
	return GetGroupByID(o, group.Id)
}

func ListGroup(o orm.Ormer, query *dataselect.DataSelectQuery) ([]Group, int64, error) {
	groups := []Group{}
	origin := o.QueryTable(Group{})
	origin, num, err := util.PaserQuerySeter(origin, nil, query, userExistKey)
	if err != nil {
		return groups, num, err
	}
	_, err = origin.All(&groups)

	return groups, num, err
}
