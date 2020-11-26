package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/golang/glog"

	"imanager/pkg/config"
	_ "imanager/pkg/db"
	"imanager/pkg/filter"
	"imanager/pkg/router"
)

func main() {
	port, err := config.GetConfig().Int(config.HttpPortKey)
	if err != nil {
		glog.Fatalf("can't get port in config")
	}
	server := &http.Server{
		Handler:      filter.GeneralFilter(router.RegisterRouter()),
		Addr:         ":" + strconv.Itoa(port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	glog.Info("Imanager Listen On " + server.Addr)
	glog.Fatal(server.ListenAndServe())
}
