package auth

import "github.com/astaxie/beego/orm"

type GroupAndRoles struct {
	Id      int `json:"id" orm:"unique"`
	GroupId int `json:"group_id"`
	RoleId  int `json:"role_id"`
}

func (g *GroupAndRoles) TableName() string {
	return "group_roles"
}

func (g *GroupAndRoles) TableUnique() [][]string {
	return [][]string{
		{"id"},
		{"group_id", "role_id"},
	}
}

func GetGroupAndRolesByGroupId(o orm.Ormer, groupId int) ([]GroupAndRoles, error) {
	res := []GroupAndRoles{}
	_, err := o.QueryTable(GroupAndRoles{}).Filter("GroupId", groupId).All(&res)
	return res, err
}

func DeleteGroupAndRolesById(o orm.Ormer, id int) error {
	groupAndRoles := GroupAndRoles{Id: id}
	_, err := o.Delete(&groupAndRoles)
	return err
}

func CreateGroupAndRoles(o orm.Ormer, roles GroupAndRoles) error {
	_, err := o.Insert(roles)
	return err
}