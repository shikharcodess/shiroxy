// Package utils provides utility functions for general-purpose use.
package utils

import (
	"fmt"
	"reflect"
)

// DestructureStruct takes a struct as input and converts its fields into a map.
// Parameters:
//   - target: interface{}, the input struct to be destructured.
//
// Returns:
//   - map[string]any: a map containing the struct's field names as keys and their corresponding values.
//
// Usage:
//
//	This function uses reflection to iterate over the fields of a struct, making it possible to convert any struct
//	(or a pointer to a struct) into a map. If a pointer is passed, it dereferences the pointer before processing.
func DestructureStruct(target interface{}) (map[string]any, error) {
	destructured := make(map[string]any)

	// Get the reflection value of the target.
	value := reflect.ValueOf(target)

	// If the input is a pointer, dereference it.
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil, fmt.Errorf("input is a nil pointer")
		}
		value = value.Elem()
	}

	// Ensure the input is a struct after potential dereferencing.
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct or pointer to struct")
	}

	// Iterate over the fields of the struct.
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		fieldValue := value.Field(i).Interface()
		destructured[field.Name] = fieldValue
	}

	return destructured, nil
}
