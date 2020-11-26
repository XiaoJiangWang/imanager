package auth

import "github.com/astaxie/beego/orm"

type UserAndRoles struct {
	Id     int `json:"id" orm:"unique"`
	UserId int `json:"user_id"`
	RoleId int `json:"role_id"`
}

func (g *UserAndRoles) TableName() string {
	return "user_roles"
}

func (g *UserAndRoles) TableUnique() [][]string {
	return [][]string{
		{"id"},
		{"user_id", "role_id"},
	}
}

func GetUserAndRolesByUserId(o orm.Ormer, userId int) ([]UserAndRoles, error) {
	res := []UserAndRoles{}
	_, err := o.QueryTable(GroupAndRoles{}).Filter("UserId", userId).All(&res)
	return res, err
}

func DeleteUserAndRolesById(o orm.Ormer, id int) error {
	userAndRoles := UserAndRoles{Id: id}
	_, err := o.Delete(&userAndRoles)
	return err
}

func CreateUserAndRoles(o orm.Ormer, roles UserAndRoles) error {
	_, err := o.Insert(roles)
	return err
}
