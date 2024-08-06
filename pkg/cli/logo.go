package cli

import (
	"fmt"
	"shiroxy/pkg/logger"
)

func PrintLogo() {
	fmt.Println("")
	logger.BluePrintln("||||||  ##### #   # # #####  ####  #   #  #   #  |||")
	logger.BluePrintln("|||||  #     #   # # #   # #    #  # #    # #  |||||")
	logger.BluePrintln("||||  ##### ##### # ##### #    #   #      #  |||||||")
	logger.BluePrintln("|||      # #   # # # #   #    #  # #    #  |||||||||")
	logger.BluePrintln("||  ##### #   # # #  #   ####  #   #  #  |||||||||||")
	fmt.Println("")

	// shiroxyFigure := figure.NewFigure("SHIROXY", "", true)
	// logger.BluePrintln(shiroxyFigure.String())
}
