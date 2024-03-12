package main

import (
	win "calendar-scrapper/internal/winlosetie"
)

func main() {
	win.InitCmd()
	win.Cmd.Execute()
}
