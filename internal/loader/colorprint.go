package loader

import (
	"fmt"
	"shiroxy/internal/loader/color"
)

func RedPrintln(content any) {
	fmt.Println(color.ColorRed, content, color.ColorReset)
}

func GreenPrintln(content any) {
	fmt.Println(color.ColorGreen, content, color.ColorReset)
}

func BluePrintln(content any) {
	fmt.Println(color.ColorBlue, content, color.ColorReset)
}
