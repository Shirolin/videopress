package main

import (
	"os"

	"videopress/internal/bootstrap"
)

func main() {
	os.Exit(bootstrap.Run("", os.Args))
}
