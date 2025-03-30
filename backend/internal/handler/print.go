package handler

import (
	"fmt"
	"reflect"
	"strings"
)

func UniversalPrint(v interface{}) string {
	var sb strings.Builder
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return "Provided value is not a struct"
	}
	sb.WriteString(fmt.Sprintf("%s {\n", typ.Name()))
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		sb.WriteString(fmt.Sprintf("  %s: %v\n", field.Name, fieldValue.Interface()))
	}
	sb.WriteString("}")
	return sb.String()
}
