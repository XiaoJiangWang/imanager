package auth

import "imanager/pkg/api/util"

type Role struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Annotation     string `json:"annotation"`
	util.BaseModel `json:",inline"`
}

type RoleList struct {
	Count int64  `json:"count"`
	Item  []Role `json:"item,omitempty"`
}
