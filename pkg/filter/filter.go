package filter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/golang/glog"

	authapi "imanager/pkg/api/auth"
	authsvc "imanager/pkg/services/auth"
	"imanager/pkg/util"
)

func GeneralFilter(h http.Handler) http.Handler {
	return logFilter(authFilter(h))
}

func logFilter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("request remote addr: %v, uri: %v, method: %v", r.RemoteAddr, r.RequestURI, r.Method)
		h.ServeHTTP(w, r)
	})
}

func authFilter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == authapi.GetTokenURL && r.Method == authapi.GetTokenMethod {
			glog.Infof("it should not verify the request which is for create token")
			h.ServeHTTP(w, r)
			return
		}

		match, _ := regexp.MatchString(authapi.InitUserURL, r.RequestURI)
		if match && r.Method == authapi.InitUserMethod {
			glog.Infof("it should not verify the request which is for init user")
			h.ServeHTTP(w, r)
			return
		}

		tokenStr := r.Header.Get(authapi.TokenHeaderKey)

		info, err := authsvc.ParseToken(tokenStr)
		if err != nil {
			glog.Errorf("parse token failed, err: %v", err)
			util.ReturnErrorResponseInResponseWriter(w, http.StatusBadRequest, fmt.Sprintf("parse token failed, %v", err))
			return
		}
		data, _ := json.Marshal(info)
		r.Header.Set(authapi.ParseInfo, string(data))
		h.ServeHTTP(w, r)
	})
}
