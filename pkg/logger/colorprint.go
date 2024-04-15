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

// ---------

func RedPrint(content any) {
	fmt.Printf(color.ColorRed, content, color.ColorReset)
}

func GreenPrint(content any) {
	fmt.Printf(color.ColorGreen, content, color.ColorReset)
}

func BluePrint(content any) {
	fmt.Printf(color.ColorBlue, content, color.ColorReset)
}

func YellowPrint(content any) {
	fmt.Printf(color.ColorYellow, content, color.ColorReset)
}

func PurplePrint(content any) {
	fmt.Printf(color.ColorPurple, content, color.ColorReset)
}

func CyanPrint(content any) {
	fmt.Printf(color.ColorCyan, content, color.ColorReset)
}
