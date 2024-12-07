package dino

import "fmt"

func Brachiosaurus(usaGalera bool) {
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

func Trex() {
	color := "\033[93m"
	nocolor := "\033[0m"

	dino := `
       __----_
      /    O  \                 ,        __   ___
      vvvvvvv  \               _|       |  | |   
      \-----,   ;             / | °  _  |  | \--\
            |    \___         \_; | | | \__; ___;
             \        \--_
            /|            \
         m-/  \             \_
           m-/ \  (     )      \
                (/      /        \
                 /     /_         \
                  \   \  (__       -__
                   \  \   \ :___      \__
                  _/  /  _/  /  \_       \___
                <<<__/ <<<___/    (________=====-_
           `

	fmt.Print(color)
	fmt.Println(dino)
	fmt.Print(nocolor)
}

func Triceraptops() {
	color := "\033[95m"
	nocolor := "\033[0m"

	dino := `
                            ,        __   ___
                           _|       |  | |   
             )            / | °  _  |  | \--.
            / )           \_; | | | \__; ___;
      -===_/   )
    ^-==( o     )__,,-~~~---,,,__
   (__         )                 ´"-_
     '---..v"")                       "--,,
            \                             ´´´>
             (|  |            |  |__,,--""""´
              |  |~--,_____,-~|  |
              |  |  |       | |  | 
              /__|\_|       /_|__| 
           `

	fmt.Print(color)
	fmt.Println(dino)
	fmt.Print(nocolor)
}

func Pterodactyl() {
	color := "\033[96m"
	nocolor := "\033[0m"

	dino := `
                             ^            ,        __   ___
                            //           _|       |  | |   
                           /o\          / | °  _  |  | \--.
                          //( \         \_; | | | \__; ___;
                         /// | \
    ___-----~~~~~~----M--V/ (  \ __--M----~~~~~-----___
_,-". , . , . , . , . ,\. ,\|  |/, ./, . , . , . , . , ."-,_
 "-, . , . , . , . , . ,\__/    \__/, . , . , . , . , . ,-"
    "--,, . , . , . , . , .(     ), . , . , . , . , .,--"
        "~-,,,,,________ . \,___,/ . ________,,,,,-~"
                        \_,/\   /\,_/    
                          \( \ / )/
                           |  V  |
                           m     m

           `

	fmt.Print(color)
	fmt.Println(dino)
	fmt.Print(nocolor)
}
