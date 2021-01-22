package auth

import "imanager/pkg/api/util"

type Group struct {
	ID             int           `json:"id"`
	Name           string        `json:"name"`
	Annotation     string        `json:"annotation"`
	Builtin        bool          `json:"builtin"`
	User           []UserInGroup `json:"user,omitempty"`
	Role           []RoleInGroup `json:"role,omitempty"`
	util.BaseModel `json:",inline"`
}

type GroupList struct {
	Count int64   `json:"count"`
	Item  []Group `json:"item,omitempty"`
}

type RoleInGroup struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Annotation string `json:"annotation"`
}

type UserInGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	UUID string `json:"uuid"`
}
