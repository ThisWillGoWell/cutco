package main

import (
	"encoding/json"
	"fmt"
	"stock-simulator-serverless/cli/execute"
	"strings"
)

func (ct*CommandTool) Execute(s string) {
	cte, err  := execute.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	result, err := cte.Exec(strings.Split(s, " "))
	if err != nil {
		fmt.Println(err)
		return
	}
	var resultString string
	switch result.(type) {
	case string:
		resultString = result.(string)
	default:
		b, _ := json.MarshalIndent(result, "", "  ")
		resultString = string(b)
	}
	fmt.Print(resultString)
}
