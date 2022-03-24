package main

import (
	"stock-simulator-serverless/cmd"
	"stock-simulator-serverless/src/seed"
)

func main() {
cmd.StartLocal(seed.Two)
}
