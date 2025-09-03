package main

import (
	"os"

	"github.com/onurbalmeida/k8s-cleanup/cmd"
)

func main() {
	code := cmd.Execute()
	os.Exit(code)
}
