package config

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/golang/glog"
)

var c Configer

// Configer defines how to get and set value from configuration raw data.
type Configer interface {
	Set(key, val string) error
	String(key string) string
	Int(key string) (int, error)
	Int64(key string) (int64, error)
	Bool(key string) (bool, error)
	Float(key string) (float64, error)
}

func InitConfig(path string) error {
	var err error
	c, err = NewDefaultConfig(path)
	return err
}

func GetConfig() Configer {
	return c
}

func init() {
	Init()
	flag.Parse()
	defer glog.Flush()
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)

	conf := os.Getenv("ConfPath")
	glog.Infof("conf path: %v", conf)

	err := InitConfig(conf)
	if err != nil {
		glog.Fatalf("init config failed, %v", err)
	}

	setValues()
}

var (
	httpPort    string
	HttpPortKey string = "httpport"

	dataSource    string
	DataSourceKey string = "datasource"

	encryptDir    string
	EncryptDirKey string = "encryptDir"
)

func setValues() {
	if httpPort != "" {
		_ = c.Set(HttpPortKey, httpPort)
	}
	if dataSource != "" {
		_ = c.Set(DataSourceKey, dataSource)
	}
	if encryptDir != "" {
		_ = c.Set(EncryptDirKey, encryptDir)
	}

	// set all env into config
	environ := os.Environ()
	for _, v := range environ {
		strs := strings.Split(v, "=")
		if len(strs) != 2 {
			continue
		}
		_ = c.Set(strs[0], strs[1])
	}
}

func Init() {
	flag.StringVar(&httpPort, "httpport", "8080", "listen port")
	flag.StringVar(&dataSource, "dataSource", "root:Eec0215@tcp(10.5.26.50:10196)/default?charset=utf8", "mysql data source")
	flag.StringVar(&encryptDir, "encryptDir", "", "the dir save master_key and pub_key")
}
