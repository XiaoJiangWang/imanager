package auth

type Role struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Annotation string `json:"annotation"`
}

type RoleList struct {
	Count int64  `json:"count"`
	Item  []Role `json:"item,omitempty"`
}