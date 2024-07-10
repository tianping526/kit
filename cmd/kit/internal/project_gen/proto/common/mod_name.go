package common

import (
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

func ModName() string {
	modBytes, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Println("failed to get module name")
		panic(err)
	}
	return modfile.ModulePath(modBytes)
}
