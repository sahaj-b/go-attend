package ui

import "fmt"

func Error(msg string) {
	fmt.Println(Red + msg + ResetStyle + "\r")
}
func Success(msg string) {
	fmt.Println(Green + msg + ResetStyle + "\r")
}
func Warn(msg string) {
	fmt.Println(Yellow + msg + ResetStyle + "\r")
}
