package cli

import (
	"fmt"
	"shiroxy/pkg/loader"
)

func PrintLogo() {
	fmt.Println("")
	loader.BluePrintln("||||||  ##### #   # # #####  ####  #   #  #   #  |||")
	loader.BluePrintln("|||||  #     #   # # #   # #    #  # #    # #  |||||")
	loader.BluePrintln("||||  ##### ##### # ##### #    #   #      #  |||||||")
	loader.BluePrintln("|||      # #   # # # #   #    #  # #    #  |||||||||")
	loader.BluePrintln("||  ##### #   # # #  #   ####  #   #  #  |||||||||||")
	fmt.Println("")
}
