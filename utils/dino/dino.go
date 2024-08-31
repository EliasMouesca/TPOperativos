package dino

import "fmt"

func dino(usa_galera bool) {
    color := "\033[92m"
    nocolor := "\033[0m"

    galera := `
                   |~|				
		   | |				`
    dino := 
    `		  _---_				
		 /   o \			
		(  ,___/			
	       /   /      ,        __   ___	
	      |   |      _|       |  | |	 
	      |   |     / | °  _  |  | \--\	
	      |   |     \_; | | | \__; ___;	
	      |   |				
	      |   |     			
	      |    \,,-~~~~~~~-,,_.		
	      |                    \_		
	      (                      \		
	       (|  |            |  |  \_	
		|  |~--,_____,-~|  |_   \	
		|  |  |       | |  | :   \__  	
		/__|\_|       /_|__|  '-____~)	
		`
	fmt.Print(color);
	if usa_galera {
        fmt.Println(galera);
    }
	fmt.Println(dino);
	fmt.Print(nocolor);
}
