package main

import (
	"fmt"
)

func printRed(msg string) {
	fmt.Println(red + msg + resetStyle + "\r")
}
func printGreen(msg string) {
	fmt.Println(green + msg + resetStyle + "\r")
}
func printYellow(msg string) {
	fmt.Println(yellow + msg + resetStyle + "\r")
}
