package auth

import (
	"net/http"
	"time"

	"imanager/pkg/db/auth"
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
	Roles     []*auth.Role `json:"roles,omitempty"`
}

const TokenHeaderKey = "X-Subject-Token"

const DefaultExpireTime = 30
const BaseDuration = time.Minute

const ParseInfo = "X-Subject-Info"

const GetTokenURL = "/v1/auth/tokens"
const GetTokenMethod = http.MethodPost

var (
	OpServiceRole RoleType = "op_service"
	AdminRole     RoleType = "admin"
	UserRole      RoleType = "user"
)

type RoleType string

var (
	roles = map[RoleType]int{
		OpServiceRole: 999,
		AdminRole:     888,
		UserRole:      1,
	}
)

func (r RoleType) IsLargerPermission(other RoleType) bool {
	return roles[r] > roles[other]
}
