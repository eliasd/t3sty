
package main

import (
  "fmt"
  "os"
)

func main() {
  fmt.Println("Hello, World!")

  indexFile, err := os.Open("./static/index.html")
  defer indexFile.Close()

  fmt.Println(err)

  switch 34 {
      case 1: {
          fmt.Println(1)
      }
      case 2: {
          fmt.Println(2)
      }
      case 34: {
          fmt.Println("yerr")
      }
      default: {
          fmt.Println("default")
      }
  }
}
