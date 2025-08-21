package main

import (
	"fmt"
	"strings"
)

func main() {
	var path string
	fmt.Scan(&path)
	new_path := strings.Split(strings.Trim(path, "/"), "/")
}
