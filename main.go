package main

import (
	"fmt"

	"github.com/ynachi/kmpass/app"
)

func main() {
	vm, err := app.New("2", "2G", "20G", "22.04", "yoa-bushit")
	if err != nil {
		fmt.Println("Failed to create VM")
	}
	vm.Create()
}
