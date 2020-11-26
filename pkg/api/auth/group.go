package auth

type Group struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Annotation string `json:"annotation"`
}

type GroupList struct {
	Count int64  `json:"count"`
	Item  []Group `json:"item,omitempty"`
}