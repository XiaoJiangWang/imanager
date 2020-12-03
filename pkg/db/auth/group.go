package auth

import (
	"fmt"

	"github.com/astaxie/beego/orm"

	"imanager/pkg/api/dataselect"
	"imanager/pkg/db/util"
)

type Group struct {
	Id             int     `json:"id" orm:"unique"`
	Name           string  `json:"name" orm:"unique"`
	Annotation     string  `json:"annotation"`
	Role           []*Role `json:"role" orm:"rel(m2m)"`
	User           []*User `orm:"reverse(many)"`
	util.BaseModel `json:",inline"`
}

var (
	groupExistKey = map[string]bool{
		"id":               true,
		"name":             true,
		"annotation":       true,
		"create_timestamp": true,
		"update_timestamp": true,
	}
	groupExistM2mForeignKey = map[string]string{
		"user__name": "user__name",
		"user__uuid": "user__uuid",
		"role__name": "role__role__name",
		"role__id":   "role__role__id",
	}
)

func GetGroupByName(o orm.Ormer, name string) (Group, error) {
	group := Group{}
	err := o.QueryTable(Group{}).Filter("name", name).One(&group)
	if err != nil {
		return group, err
	}
	_, err = o.LoadRelated(&group, "user")
	if err != nil {
		return group, err
	}
	_, err = o.LoadRelated(&group, "role")
	if err != nil {
		return group, err
	}
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

func UpdateGroup(o orm.Ormer, group Group) (Group, error) {
	var err error
	var oldGroup Group
	if group.Id != 0 {
		oldGroup, err = GetGroupByID(o, group.Id)
	} else if group.Name != "" {
		oldGroup, err = GetGroupByName(o, group.Name)
	} else {
		return group, fmt.Errorf("find group by name or id failed")
	}
	if err != nil {
		return group, err
	}
	err = patch(&oldGroup, &group)
	if err != nil {
		return group, err
	}

	_, err = o.Update(&group)
	if err != nil {
		return group, err
	}
	return GetGroupByName(o, group.Name)
}

func DeleteGroupByName(o orm.Ormer, name string) error {
	var err error
	if err = o.Begin(); err != nil {
		return err
	}
	group, err := GetGroupByName(o, name)
	if err != nil {
		_ = o.Rollback()
		return err
	}
	m2m := o.QueryM2M(&group, "role")
	for _, v := range group.Role {
		if _, err = m2m.Remove(v); err != nil {
			_ = o.Rollback()
			return err
		}
	}
	if _, err = o.Delete(&group); err != nil {
		_ = o.Rollback()
		return err
	}
	_ = o.Commit()
	return nil
}

func CreateGroup(o orm.Ormer, group Group) (Group, error) {
	_, err := o.Insert(&group)
	if err != nil {
		return group, err
	}
	return GetGroupByName(o, group.Name)
}

func ListGroup(o orm.Ormer, query *dataselect.DataSelectQuery) ([]Group, int64, error) {
	groups := []Group{}
	origin := o.QueryTable(Group{})
	origin, num, err := util.PaserQuerySeter(origin, nil, query, groupExistKey, groupExistM2mForeignKey)
	if err != nil {
		return groups, num, err
	}
	_, err = origin.All(&groups)

	return groups, num, err
}
