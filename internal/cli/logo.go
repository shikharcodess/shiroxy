package cli

import (
	"fmt"
	"shiroxy/internal/loader"
)

func PrintLogo() {
	fmt.Println("")
	loader.BluePrintln("||      ##### #   # # #####  ####  #   #  #   #  ||AVRAM||")
	loader.BluePrintln("||     #     #   # # #   # #    #  # #    # #    ||19:45||")
	loader.BluePrintln("||    ##### ##### # ##### #    #   #      #      |\\//\\|\\||")
	loader.BluePrintln("||       # #   # # # #   #    #  # #    #        |||||||||")
	loader.BluePrintln("||  ##### #   # # #  #   ####  #   #  #          |||||||||")
	fmt.Println("")
}
