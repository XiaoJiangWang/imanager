package filter

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"

	authapi "imanager/pkg/api/auth"
	authsvc "imanager/pkg/services/auth"
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
			h.ServeHTTP(w,r)
			return
		}

		tokenStr := r.Header.Get(authapi.TokenHeaderKey)

		info, err := authsvc.ParseToken(tokenStr)
		if err != nil {
			glog.Errorf("parse token failed, err: %v", err)
			// TODO: should exit
			h.ServeHTTP(w, r)
			return
		}
		data, _ := json.Marshal(info)
		r.Header.Set(authapi.ParseInfo, string(data))
		h.ServeHTTP(w, r)
	})
}