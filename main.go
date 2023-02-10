package main

import (
	"fmt"

	"github.com/ynachi/kmpass/app"
)

func main() {
	vm, err := app.New("2", "2G", "20G", "20.04", "yoa-bushit", "/home/ynachi/codes/github.com/kmpass2/app/files/clouds.yaml")
	if err != nil {
		fmt.Println("Failed to create VM")
	}
	vm.Create()
}
