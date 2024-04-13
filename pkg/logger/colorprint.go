package logger

import (
	"fmt"
	"shiroxy/pkg/color"
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

func YellowPrintln(content any) {
	fmt.Println(color.ColorYellow, content, color.ColorReset)
}

func PurplePrintln(content any) {
	fmt.Println(color.ColorPurple, content, color.ColorReset)
}

func CyanPrintln(content any) {
	fmt.Println(color.ColorCyan, content, color.ColorReset)
}
