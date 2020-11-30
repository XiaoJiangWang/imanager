package auth

import "imanager/pkg/api/util"

type User struct {
	UUID           string       `json:"uuid"`
	Name           string       `json:"name"`
	Password       string       `json:"password,omitempty"`
	TruthName      string       `json:"truth_name"`
	Email          string       `json:"email"`
	PhoneNum       string       `json:"phone_num"`
	Group          *GroupInUser `json:"group"`
	Role           []RoleInUser `json:"role"`
	util.BaseModel `json:",inline"`
}

type RoleInUser struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Annotation string `json:"annotation"`
}

type GroupInUser struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Annotation string `json:"annotation"`
}

type UserList struct {
	Count int64  `json:"count"`
	Item  []User `json:"item,omitempty"`
}
