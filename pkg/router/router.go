package router

import (
	"net/http"

	"github.com/gorilla/mux"

	"imanager/pkg/controllers"
)

func RegisterRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/v1/auth/tokens", controllers.AuthController{}.CreateTokenInHttp).Methods(http.MethodPost)
	r.HandleFunc("/v1/auth/tokens", controllers.AuthController{}.CheckTokenInHttp).Methods(http.MethodGet)

	r.HandleFunc("/v1/auth/user", controllers.AuthController{}.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/v1/auth/user", controllers.AuthController{}.ModifyUser).Methods(http.MethodPut)
	r.HandleFunc("/v1/auth/user/{name}", controllers.AuthController{}.DeleteUser).Methods(http.MethodDelete)
	r.HandleFunc("/v1/auth/user", controllers.AuthController{}.ListUser).Methods(http.MethodGet)
	r.HandleFunc("/v1/auth/user/{name}", controllers.AuthController{}.GetUser).Methods(http.MethodGet)

	r.HandleFunc("/v1/auth/role", controllers.AuthController{}.CreateRole).Methods(http.MethodPost)
	r.HandleFunc("/v1/auth/role", controllers.AuthController{}.ModifyRole).Methods(http.MethodPut)
	r.HandleFunc("/v1/auth/role/{name}", controllers.AuthController{}.DeleteRole).Methods(http.MethodDelete)
	r.HandleFunc("/v1/auth/role", controllers.AuthController{}.ListRole).Methods(http.MethodGet)
	r.HandleFunc("/v1/auth/role/{name}", controllers.AuthController{}.GetRole).Methods(http.MethodGet)

	r.HandleFunc("/v1/auth/group", controllers.AuthController{}.CreateGroup).Methods(http.MethodPost)
	r.HandleFunc("/v1/auth/group", controllers.AuthController{}.ModifyGroup).Methods(http.MethodPut)
	r.HandleFunc("/v1/auth/group/{name}", controllers.AuthController{}.DeleteGroup).Methods(http.MethodDelete)
	r.HandleFunc("/v1/auth/group", controllers.AuthController{}.ListGroup).Methods(http.MethodGet)
	r.HandleFunc("/v1/auth/group/{name}", controllers.AuthController{}.GetGroup).Methods(http.MethodGet)

	r.HandleFunc("/v1/auth/user/{name}/init", controllers.AuthController{}.InitUser).Methods(http.MethodPut)
	r.HandleFunc("/v1/auth/user/{name}/uninit", controllers.AuthController{}.UnInitUser).Methods(http.MethodPut)
	return r
}