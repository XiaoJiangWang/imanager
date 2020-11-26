package util

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"

	"imanager/pkg/api"
)

func ReturnErrorResponseInResponseWriter(w http.ResponseWriter, errorCode int, errorMessage string) {
	glog.Errorf("return core: %v, error message: %v", errorCode, errorMessage)
	w.WriteHeader(errorCode)
	tmp := api.ErrorResponse{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
	data, _ := json.Marshal(tmp)
	w.Write(data)
	return
}