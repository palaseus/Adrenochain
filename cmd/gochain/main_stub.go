//go:build !go1.20
// +build !go1.20

package main

import "fmt"

func main() {
	fmt.Println("adrenochain CLI requires Go 1.20+. Please build with a newer Go version.")
}
