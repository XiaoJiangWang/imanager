package auth

import "time"

type UserInHarbor struct {
	Username        string    `json:"username"`
	Comment         string    `json:"comment"`
	UpdateTime      time.Time `json:"update_time"`
	Password        string    `json:"password"`
	UserID          int       `json:"user_id"`
	RealName        string    `json:"realname"`
	Deleted         bool      `json:"deleted"`
	CreationTime    time.Time `json:"creation_time"`
	AdminRoleInAuth bool      `json:"admin_role_in_auth"`
	RoleID          int       `json:"role_id"`
	SysadminFlag    bool      `json:"sysadmin_flag"` // 是否为管理员
	RoleName        string    `json:"role_name"`
	ResetUUID       string    `json:"reset_uuid"`
	Salt            string    `json:"salt"`
	Email           string    `json:"email"`
}

type TokenInHarbor struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type ErrorMessageInHarbor struct {
	Errors []BaseErrorMessageInHarbor `json:"errors"`
}

type BaseErrorMessageInHarbor struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type UserUpdateProfileInHarbor struct {
	Comment  string `json:"comment"`
	Email    string `json:"email"`
	RealName string `json:"realname"`
}

type UserUpdatePasswordInHarbor struct {
	NewPassword string `json:"new_password"`
	OldPassword string `json:"old_password"`
}

type UserUpdateSysAdminInHarbor struct {
	SysadminFlag bool `json:"sysadmin_flag"` // 是否为管理员
}

type UserSearchInHarbor struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
}

type RespForSearchUserIDInHarbor []BaseSearchUserIDInHarbor

type BaseSearchUserIDInHarbor struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}
