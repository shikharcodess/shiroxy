package utils

import (
	"encoding/json"
	"fmt"
)

func LogStruct(s interface{}) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Println("Error logging struct: ", err.Error())
	}

	fmt.Println(string(data))
}
