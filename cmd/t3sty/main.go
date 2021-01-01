package main

import (
  "fmt"
  "github.com/eliasd/t3sty/pkg/t3sty"
)

// Main file executable should be run from the ./t3sty/ directory.
//  - that is it's path is ./t3sty/main
func main() {
  fmt.Println("Starting server...")
  t3sty.StartServer()
}
