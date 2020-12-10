package util

import (
	"errors"
	"fmt"
	"reflect"
)

func Patch(old, new interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintf("%v", e))
		}
	}()

	oldType, oldValue := reflect.TypeOf(old), reflect.ValueOf(old)
	newType, newValue := reflect.TypeOf(new), reflect.ValueOf(new)

	if oldType.Kind() != reflect.Ptr || newType.Kind() != reflect.Ptr ||
		oldType.Elem().Kind() != reflect.Struct || newType.Elem().Kind() != reflect.Struct {
		return errors.New("type should be a struct pointer")
	}
	oldType, oldValue, newType, newValue = oldType.Elem(), oldValue.Elem(), newType.Elem(), newValue.Elem()

	for i := 0; i < oldType.NumField(); i++ {
		property := oldType.Field(i)
		oldPropertyValue := oldValue.FieldByName(property.Name)
		if newValue.FieldByName(property.Name).IsZero() && newValue.Field(i).CanSet() {
			newValue.Field(i).Set(oldPropertyValue)
		}
	}
	return nil
}