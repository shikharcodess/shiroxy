package utils

import "reflect"

func DestructureStruct(target interface{}) map[string]any {
	destructured := map[string]any{}

	analyticsReflectValue := reflect.ValueOf(target)
	if analyticsReflectValue.Kind() == reflect.Ptr {
		analyticsReflectValue = analyticsReflectValue.Elem()
	}
	for i := 0; i < analyticsReflectValue.NumField(); i++ {
		destructured[analyticsReflectValue.Type().Field(i).Name] = analyticsReflectValue.Field(i).Interface()
	}

	return destructured
}
