package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
)

type CommandTool struct {
}

// pdms opps cli tool
func main() {
	args := os.Args

	ct := &CommandTool{}

	// if there are no args, then just do a single prompt
	if len(args) == 1 {
		t := prompt.Input("> ", ct.Complete)
		fmt.Println(t)
		ct.Execute(t)
		return
	}
	// if there are more, run the command as normal
	ct.Execute(strings.Join(args[1:], " "))

}

func list() {

}
