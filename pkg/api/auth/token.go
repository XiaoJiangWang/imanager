package auth

import (
	"net/http"
	"time"
)

type ReqToken struct {
	Auth  ReqTokenAuth  `json:"auth,omitempty"`
	Scope ReqTokenScope `json:"scope,omitempty"`
}

type ReqTokenAuth struct {
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
}

type ReqTokenScope struct {
	Duration time.Duration `json:"duration,omitempty"`
}

type RespToken struct {
	ExpiresAt time.Time    `json:"expires_at,omitempty"`
	IssuedAt  time.Time    `json:"issued_at,omitempty"`
	UserID    string       `json:"user_id"`
	Name      string       `json:"name,omitempty"`
	TrueName  string       `json:"true_name,omitempty"`
	Group     *GroupInUser `json:"group,omitempty"`
	Role      []RoleInUser `json:"roles,omitempty"`
}

const TokenHeaderKey = "X-Subject-Token"

const DefaultExpireTime = 30
const BaseDuration = time.Minute

const ParseInfo = "X-Subject-Info"

const GetTokenURL = "/v1/auth/tokens"
const GetTokenMethod = http.MethodPost

var (
	OpServiceRole RoleType = 1
	AdminRole     RoleType = 2
	UserRole      RoleType = 3
)

type RoleType int

func (r RoleType) String() string {
	str, ok := rolesString[r]
	if !ok {
		return "invalid role type"
	}
	return str
}

var (
	roles = map[RoleType]int{
		OpServiceRole: 999,
		AdminRole:     888,
		UserRole:      1,
	}
	rolesString = map[RoleType]string {
		OpServiceRole: "op_service",
		AdminRole:     "admin",
		UserRole:      "user",
	}
)

func (r RoleType) IsLargerPermission(other RoleType) bool {
	return roles[r] >= roles[other]
}

func GetLargestRolePermission(role []RoleInUser) RoleType {
	res := UserRole
	if role == nil {
		return res
	}
	for _, v := range role {
		tmp := RoleType(v.ID)
		if tmp.IsLargerPermission(res) {
			res = tmp
		}
	}
	return res
}