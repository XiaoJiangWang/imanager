package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
)

type DefaultConfig struct {
	path string
	data map[string]string
	sync.RWMutex
}

func NewDefaultConfig(path string) (*DefaultConfig, error) {
	var err error
	tmp := &DefaultConfig{
		path:    path,
		data:    make(map[string]string),
		RWMutex: sync.RWMutex{},
	}
	tmp.data, err = readFile(path)
	return tmp, err
}

func readFile(path string) (map[string]string, error) {
	if path == "" {
		return make(map[string]string), nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	buf := bufio.NewReader(bytes.NewBuffer(data))
	for {
		tmp, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, err
		}

		tmp = bytes.TrimSpace(tmp)
		if len(tmp) == 0 {
			continue
		}
		if tmp[0] == '#' || strings.HasPrefix(string(tmp), "//") {
			// ignore annotation
			continue
		}
		keyValue := bytes.SplitN(tmp, []byte("="), 2)
		if len(keyValue) != 2 {
			return nil, errors.New("read the content error: \"" + string(tmp) + "\", should key = val")
		}

		key := string(bytes.TrimSpace(keyValue[0])) // key name case insensitive
		key = strings.ToLower(key)

		val := bytes.TrimSpace(keyValue[1])
		if bytes.HasPrefix(val, []byte(`"`)) {
			val = bytes.Trim(val, `"`)
		}

		res[key] = string(val)
	}
	return res, nil
}

func (c *DefaultConfig) Set(key, val string) error {
	if len(key) == 0 {
		return errors.New("key is empty")
	}
	key = strings.ToLower(key)
	c.Lock()
	defer c.Unlock()
	c.data[key] = val
	return nil
}

func (c *DefaultConfig) get(key string) string {
	if len(key) == 0 {
		return ""
	}
	key = strings.ToLower(key)
	c.RLock()
	defer c.RUnlock()
	return c.data[key]
}

func (c *DefaultConfig) String(key string) string {
	return c.get(key)
}

func (c *DefaultConfig) Int(key string) (int, error) {
	return strconv.Atoi(c.String(key))
}

func (c *DefaultConfig) Int64(key string) (int64, error) {
	return strconv.ParseInt(c.String(key), 10, 64)
}

func (c *DefaultConfig) Bool(key string) (bool, error) {
	return parseBool(c.get(key))
}

func (c *DefaultConfig) Float(key string) (float64, error) {
	return strconv.ParseFloat(c.get(key), 64)
}

func parseBool(val interface{}) (value bool, err error) {
	if val != nil {
		switch v := val.(type) {
		case bool:
			return v, nil
		case string:
			switch v {
			case "1", "t", "T", "true", "TRUE", "True", "YES", "yes", "Yes", "Y", "y", "ON", "on", "On":
				return true, nil
			case "0", "f", "F", "false", "FALSE", "False", "NO", "no", "No", "N", "n", "OFF", "off", "Off":
				return false, nil
			}
		case int8, int32, int64:
			strV := fmt.Sprintf("%d", v)
			if strV == "1" {
				return true, nil
			} else if strV == "0" {
				return false, nil
			}
		case float64:
			if v == 1.0 {
				return true, nil
			} else if v == 0.0 {
				return false, nil
			}
		}
		return false, fmt.Errorf("parsing %q: invalid syntax", val)
	}
	return false, fmt.Errorf("parsing <nil>: invalid syntax")
}
