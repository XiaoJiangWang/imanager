package auth

import (
	"fmt"

	"github.com/astaxie/beego/orm"

	"imanager/pkg/api/dataselect"
	"imanager/pkg/db/util"
)

type Role struct {
	Id             int      `json:"id" orm:"unique"`
	Name           string   `json:"name" orm:"unique"`
	Annotation     string   `json:"annotation"`
	User           []*User  `json:"-" orm:"reverse(many)"`
	Group          []*Group `json:"-" orm:"reverse(many)"`
	util.BaseModel `json:",inline"`
}

var (
	roleExistKey = map[string]bool{
		"id":               true,
		"name":             true,
		"annotation":       true,
		"create_timestamp": true,
		"update_timestamp": true,
	}
	roleExistM2mForeignKey = map[string]string{
		"user__name":  "user__user__name",
		"user__uuid":  "user__user__uuid",
		"group__id":   "group__group__id",
		"group__name": "group__group__name",
	}
)

func GetRoleByName(o orm.Ormer, name string) (Role, error) {
	role := Role{}
	err := o.QueryTable(Role{}).Filter("name", name).One(&role)
	if err != nil {
		return role, err
	}
	_, err = o.LoadRelated(&role, "user")
	if err != nil {
		return role, err
	}
	_, err = o.LoadRelated(&role, "group")
	if err != nil {
		return role, err
	}
	return role, err
}

func GetRoleByID(o orm.Ormer, id int) (Role, error) {
	role := Role{}
	err := o.QueryTable(Role{}).Filter("id", id).One(&role)
	_, err = o.LoadRelated(&role, "user")
	if err != nil {
		return role, err
	}
	_, err = o.LoadRelated(&role, "group")
	if err != nil {
		return role, err
	}
	return role, err
}

func UpdateRole(o orm.Ormer, role Role) (Role, error) {
	var err error
	var oldRole Role
	if role.Id != 0 {
		oldRole, err = GetRoleByID(o, role.Id)
	} else if role.Name != "" {
		oldRole, err = GetRoleByName(o, role.Name)
	} else {
		return role, fmt.Errorf("find role by name or id failed")
	}
	if err != nil {
		return role, err
	}
	err = patch(&oldRole, &role)
	if err != nil {
		return role, err
	}

	_, err = o.Update(&role)
	if err != nil {
		return role, err
	}
	return GetRoleByName(o, role.Name)
}

func DeleteRoleByName(o orm.Ormer, name string) error {
	role, err := GetRoleByName(o, name)
	if err != nil {
		return err
	}
	_, err = o.Delete(&role)
	return err
}

func CreateRole(o orm.Ormer, role Role) (Role, error) {
	_, err := o.Insert(&role)
	if err != nil {
		return role, err
	}
	return GetRoleByName(o, role.Name)
}

func ListRole(o orm.Ormer, query *dataselect.DataSelectQuery) ([]Role, int64, error) {
	roles := []Role{}
	origin := o.QueryTable(Role{})
	origin, num, err := util.ParseQuerySeter(origin, nil, query, roleExistKey, roleExistM2mForeignKey)
	if err != nil {
		return roles, num, err
	}
	_, err = origin.All(&roles)
	return roles, num, err
}
