package logger

import "fmt"

func PrintLogo() {
	fmt.Println("    ##### #   # # #####  ####  #   #  #   #")
	fmt.Println("   #     #   # # #   # #    #  # #    # #")
	fmt.Println("  ##### ##### # ##### #    #   #      #")
	fmt.Println("     # #   # # # #   #    #  # #    #")
	fmt.Println("##### #   # # #  #   ####  #   #  #")
}

//  /items/email_marketing?filter={publish_on:{_between:[2024-1-11T14:48:0,2024-1-11T14:50:0]},status:{_eq:active},type:{_eq:email_journey}}&fields=*,customers.customer_id.*,template.*&
