package main

import (
	"os"

	"github.com/alexppxela/order-gatling/cmd"
)

func main() {
	err := cmd.OrderGatlingCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
