package auth

import (
	"github.com/astaxie/beego/orm"

	"imanager/pkg/api/dataselect"
	"imanager/pkg/db/util"
)

type Role struct {
	Id         int      `json:"id" orm:"unique"`
	Name       string   `json:"name" orm:"unique"`
	Annotation string   `json:"annotation"`
	User       []*User  `json:"-" orm:"reverse(many)"`
	Group      []*Group `json:"-" orm:"reverse(many)"`
}

func GetRoleByName(o orm.Ormer, name string) (Role, error) {
	role := Role{}
	err := o.QueryTable(Role{}).Filter("name", name).One(&role)
	return role, err
}

func GetRoleByID(o orm.Ormer, id int) (Role, error) {
	role := Role{}
	err := o.QueryTable(Role{}).Filter("id", id).One(&role)
	return role, err
}

func UpdateRole(o orm.Ormer, role Role) error {
	_, err := o.Update(&role)
	return err
}

func DeleteRoleByName(o orm.Ormer, name string) error {
	role := Role{Name: name}
	_, err := o.Delete(&role)
	return err
}

func CreateRole(o orm.Ormer, role Role) (Role, error) {
	_, err := o.Insert(role)
	if err != nil {
		return role, err
	}
	return GetRoleByID(o, role.Id)
}

func ListRole(o orm.Ormer, query *dataselect.DataSelectQuery) ([]Role, int64, error) {
	roles := []Role{}
	origin := o.QueryTable(Role{})
	origin, num, err := util.PaserQuerySeter(origin, nil, query, userExistKey)
	if err != nil {
		return roles, num, err
	}
	_, err = origin.All(&roles)
	return roles, num, err
}
