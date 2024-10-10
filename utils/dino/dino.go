package dino

import "fmt"

func Dino(usaGalera bool) {
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
	fmt.Print(color)
	if usaGalera {
		fmt.Println(galera)
	}
	fmt.Println(dino)
	fmt.Print(nocolor)
}
