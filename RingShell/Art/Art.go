package Art

import (
	"fmt"

	"github.com/common-nighthawk/go-figure"
)

func Art() {
	// Define ANSI color codes
	reset := "\033[0m"
	cyan := "\033[36m"
	green := "\033[32m"
	red := "\033[31m"
	grey := "\033[90m" // Light grey for developer info

	// Title: RingShell in ASCII art
	title := figure.NewFigure("RingShell", "", true)
	fmt.Print(cyan)
	title.Print()
	fmt.Print(reset)

	// Developer info displayed as a list
	fmt.Println()
	fmt.Println(grey + "Developed by: " + green + "Mustafa Abdulaziz Hamza" + reset)
	fmt.Println(grey + "GitHub: " + red + "github.com/MustafaAbdulazizHamza" + reset)
	fmt.Println(grey + "------------------------------------------" + reset)

	// Poem from Lord of the Rings
	poem := `
		` + green + `One Ring to rule them all, 
		One Ring to find them,
		One Ring to bring them all 
		and in the darkness bind them.
		In the Land of Mordor 
		where the Shadows lie.` + reset

	fmt.Println(poem)
}
