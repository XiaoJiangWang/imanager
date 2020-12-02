package auth

import (
	"fmt"

	"github.com/astaxie/beego/orm"

	"imanager/pkg/api/dataselect"
	"imanager/pkg/db/util"
)

type User struct {
	ID             int     `json:"id" orm:"column(id);unique"`
	UUID           string  `json:"uuid" orm:"column(uuid);unique"`
	Name           string  `json:"name" orm:"unique"`
	Password       string  `json:"password" orm:"type(text)"`
	Role           []*Role `json:"role" orm:"rel(m2m)"`
	TruthName      string  `json:"truthname"`
	Email          string  `json:"email"`
	PhoneNum       string  `json:"phonenum"`
	Group          *Group  `json:"group" orm:"rel(fk)"`
	util.BaseModel `json:",inline"`
}

var userExistKey = map[string]bool{
	"id":               true,
	"uuid":             true,
	"name":             true,
	"truth_name":       true,
	"email":            true,
	"phone_num":        true,
	"create_timestamp": true,
	"update_timestamp": true,
}

func GetUserByName(o orm.Ormer, name string) (User, error) {
	user := User{}
	err := o.QueryTable(User{}).Filter("name", name).One(&user)
	if err != nil {
		return User{}, err
	}
	_, err = o.LoadRelated(&user, "role")
	if err != nil {
		return user, err
	}
	_, err = o.LoadRelated(&user, "group")
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetUserByUUID(o orm.Ormer, uuid string) (User, error) {
	user := User{}
	err := o.QueryTable(User{}).Filter("uuid", uuid).One(&user)
	if err != nil {
		return User{}, err
	}
	_, err = o.LoadRelated(&user, "role")
	if err != nil {
		return user, err
	}
	_, err = o.LoadRelated(&user, "group")
	if err != nil {
		return user, err
	}
	return user, nil
}

func UpdateUser(o orm.Ormer, user User) (User, error) {
	var err error
	var oldUser User
	if user.UUID != "" {
		oldUser, err = GetUserByUUID(o, user.UUID)
	} else if user.Name != "" {
		oldUser, err = GetUserByName(o, user.Name)
	} else {
		return user, fmt.Errorf("find user by name or uuid failed")
	}
	if err != nil {
		return user, err
	}
	err = patch(&oldUser, &user)
	if err != nil {
		return user, err
	}

	err = o.Begin()
	if err != nil {
		return user, err
	}
	_, err = o.Update(&user)
	if err != nil {
		_ = o.Rollback()
		return user, err
	}
	if HasDifferentRole(oldUser, user) {
		m2m := o.QueryM2M(&user, "role")
		if _, err = m2m.Clear(); err != nil {
			_ = o.Rollback()
			return user, err
		}
		for _, v := range user.Role {
			if m2m.Exist(v) {
				continue
			}
			_, err = m2m.Add(v)
			if err != nil {
				_ = o.Rollback()
				return user, err
			}
		}
	}
	user, err = GetUserByName(o, user.Name)
	if err != nil {
		_ = o.Rollback()
		return user, err
	}
	_ = o.Commit()

	return user, err
}

func HasDifferentRole(oldUser, user User) bool {
	if len(oldUser.Role) != len(user.Role) {
		return true
	}
	m1 := make(map[string]bool, len(oldUser.Role))
	for _, r := range oldUser.Role {
		m1[r.Name] = true
	}
	for _, r := range user.Role {
		if !m1[r.Name] {
			return true
		}
	}
	return false
}

func DeleteUserByName(o orm.Ormer, name string) error {
	var err error
	err = o.Begin()
	if err != nil {
		return err
	}
	user, err := GetUserByName(o, name)
	if err != nil {
		_ = o.Rollback()
		return err
	}
	m2m := o.QueryM2M(&user, "role")
	if _, err = m2m.Clear(); err != nil {
		_ = o.Rollback()
		return err
	}
	_, err = o.Delete(&user)
	if err != nil {
		_ = o.Rollback()
		return err
	}
	_ = o.Commit()
	return nil
}

func CreateUser(o orm.Ormer, user User) (User, error) {
	err := o.Begin()
	if err != nil {
		return user, err
	}
	_, err = o.Insert(&user)
	if err != nil {
		_ = o.Rollback()
		return user, err
	}

	m2m := o.QueryM2M(&user, "role")
	for _, v := range user.Role {
		if m2m.Exist(v) {
			continue
		}
		_, err = m2m.Add(v)
		if err != nil {
			_ = o.Rollback()
			return user, err
		}
	}

	user, err = GetUserByName(o, user.Name)
	if err != nil {
		_ = o.Rollback()
		return user, err
	}
	_ = o.Commit()
	return user, err
}

func ListUsersByUserIDs(o orm.Ormer, userIDs []string, query *dataselect.DataSelectQuery) ([]User, int64, error) {
	users := []User{}
	origin := o.QueryTable(User{})
	origin, num, err := util.PaserQuerySeter(origin, userIDs, query, userExistKey)
	if err != nil {
		return users, 0, err
	}
	_, err = origin.All(&users)
	for k := range users {
		_, err = o.LoadRelated(&users[k], "role")
		if err != nil {
			return users, num, err
		}
		_, err = o.LoadRelated(&users[k], "group")
		if err != nil {
			return users, num, err
		}
	}

	return users, num, err
}
