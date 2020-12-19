package util

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const (
	arrayPrefix = "["
	arraySuffix = "]"
	arrayLink   = "=>"
	arrayEnding = ","
)

var deep = 0

// GoTypeToPHPCode 将Go简单数据类型转换为PHP数组
func GoTypeToPHPCode(v interface{}) string {
	t := reflect.TypeOf(v)
	value := reflect.ValueOf(v)
	switch t.Kind() {
	case reflect.Map:
		return isMap(value)
	case reflect.Array, reflect.Slice:
		return isSlice(value)
	case reflect.Struct:
		return isStruct(t, value)
	case reflect.String:
		return isString(value)
	default:
		return isOther(value)
	}
}

func tab() (str string) {
	str = ""
	for i := 0; i < deep; i++ {
		str += "\t"
	}
	return
}

func isMap(v reflect.Value) string {
	deep++
	str := arrayPrefix + "\n"
	keys := v.MapKeys()
	sortKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		sortKeys = append(sortKeys, key.Interface().(string))
	}
	sort.Strings(sortKeys)
	for _, k := range sortKeys {
		key := reflect.ValueOf(k)
		value := v.MapIndex(key)
		str += fmt.Sprintf("%s%v %s %v%s\n",
			tab(), GoTypeToPHPCode(key.Interface()), arrayLink, GoTypeToPHPCode(value.Interface()), arrayEnding)
	}
	deep--
	str += tab() + arraySuffix
	return str
}

func isSlice(v reflect.Value) string {
	deep++
	str := arrayPrefix + "\n"
	for i := 0; i < v.Len(); i++ {
		str += fmt.Sprintf("%s%v%s\n",
			tab(), GoTypeToPHPCode(v.Index(i).Interface()), arrayEnding)
	}
	deep--
	str += tab() + arraySuffix
	return str
}

func isStruct(t reflect.Type, v reflect.Value) string {
	deep++
	str := arrayPrefix + "\n"
	for i := 0; i < v.NumField(); i++ {
		key := t.Field(i).Tag.Get("php")
		str += fmt.Sprintf("%s'%s' %s %v%s\n",
			tab(), key, arrayLink, GoTypeToPHPCode(v.Field(i).Interface()), arrayEnding)
	}
	deep--
	str += tab() + arraySuffix
	return str
}

func isString(v reflect.Value) string {
	return fmt.Sprintf("'%s'", strings.Replace(v.String(), "'", "\\'", -1))
}

func isOther(v reflect.Value) string {
	return fmt.Sprintf("%v", v.Interface())
}
