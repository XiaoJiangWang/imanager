package db

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"

	"imanager/pkg/config"
	"imanager/pkg/db/auth"
)

func init() {
	err := orm.RegisterDriver("mysql", orm.DRMySQL)
	if err != nil {
		panic(fmt.Sprintf("register database driver failed, err: %v", err))
	}

	dataSource := config.GetConfig().String("dataSource")
	glog.Infof("datasource: %v", dataSource)
	err = orm.RegisterDataBase("default", "mysql", dataSource)
	if err != nil {
		panic(fmt.Sprintf("register database failed, err: %v", err))
	}

	orm.SetMaxOpenConns("default", 30)
	orm.SetMaxIdleConns("default", 30)
	orm.DefaultTimeLoc = time.UTC

	orm.RegisterModel(new(auth.User), new(auth.Role), new(auth.Group))

	err = orm.RunSyncdb("default", false, false)
	if err != nil {
		panic(fmt.Sprintf("create database table failed, err: %v", err))
	}
}