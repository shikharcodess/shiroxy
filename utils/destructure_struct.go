// Package utils provides utility functions for general-purpose use.
package utils

import "reflect"

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
func DestructureStruct(target interface{}) map[string]any {
	destructured := map[string]any{} // Initialize an empty map to store the struct's fields.

	// Use reflection to get the value of the input struct.
	analyticsReflectValue := reflect.ValueOf(target)
	if analyticsReflectValue.Kind() == reflect.Ptr {
		// If the input is a pointer, get the element it points to.
		analyticsReflectValue = analyticsReflectValue.Elem()
	}

	// Iterate over the fields of the struct.
	for i := 0; i < analyticsReflectValue.NumField(); i++ {
		// Add each field's name and value to the map.
		destructured[analyticsReflectValue.Type().Field(i).Name] = analyticsReflectValue.Field(i).Interface()
	}

	return destructured
}
