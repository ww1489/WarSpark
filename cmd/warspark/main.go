package main

import (
	"os"
)

// @title WarSpark API
// @version dev
// @description Reusable Go REST API backend template.
// @BasePath /
// @schemes http
func main() {
	if err := execute(); err != nil {
		os.Exit(1)
	}
}
