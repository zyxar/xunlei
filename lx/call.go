package main

import (
	"reflect"
)

func Call(name interface{}, params []string) (result []reflect.Value, err error) {
	fn := reflect.ValueOf(name)
	if len(params) != fn.Type().NumIn() {
		err = errInvalidArgs
		return
	}
	args := make([]reflect.Value, len(params))
	for k, param := range params {
		args[k] = reflect.ValueOf(param)
	}
	result = fn.Call(args)
	return
}
