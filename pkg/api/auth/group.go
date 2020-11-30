package auth

import "imanager/pkg/api/util"

type Group struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Annotation     string `json:"annotation"`
	util.BaseModel `json:",inline"`
}

type GroupList struct {
	Count int64   `json:"count"`
	Item  []Group `json:"item,omitempty"`
}
