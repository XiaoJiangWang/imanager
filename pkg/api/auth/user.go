package auth

import (
	"net/http"

	"imanager/pkg/api/util"
)

const InitUserURL = "/v1/auth/user/[a-zA-Z0-9-]{4,64}/init"
const InitUserMethod = http.MethodPut

type User struct {
	UUID           string       `json:"uuid"`
	Name           string       `json:"name"`
	Password       string       `json:"-"`
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

type UserSecret struct {
	Password string `json:"password"`
}
