package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/dyatlov/gostardict/stardict"
)

func main() {
	dictPath := flag.String("p", "", "Specify directory where dictionary files are located")
	dictName := flag.String("n", "", "Specify name of dictionary")
	flag.Parse()

	// init dictionary with path to dictionary files and name of dictionary
	dict, err := stardict.NewDictionary(*dictPath, *dictName)

	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(" > ")
		item, _ := reader.ReadString('\n') // Prompt user for a word to translate

		item = item[:len(item)-1]

		if item == "" { // if user entered nothing - then finish
			break
		}

		senses := dict.Translate(item) // get translations

		for i, seq := range senses { // for each translation analyze returned parts
			fmt.Printf("Sense %d\n", i+1)
			for j, t := range seq.Parts { // write each part contents to user
				fmt.Printf("Part %d:\n%c\n%s\n", j+1, t.Type, t.Data)
			}
		}

		if len(senses) == 0 {
			fmt.Println("Nothing found :(")
		}
	}
}
