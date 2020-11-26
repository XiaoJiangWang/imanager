package util

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"

	authapi "imanager/pkg/api/auth"
)

func GetUserInfo(data string) (*authapi.RespToken, error) {
	if len(data) == 0 {
		glog.Errorf("user info is empty, token is expired or empty.")
		return nil, fmt.Errorf("user info is empty, token is expired or empty")
	}

	info := authapi.RespToken{}
	err := json.Unmarshal([]byte(data), &info)
	if err != nil {
		glog.Errorf("user info read failed, %v", err)
		return nil, fmt.Errorf("user info read failed, %v", err)
	}
	return &info, nil
}