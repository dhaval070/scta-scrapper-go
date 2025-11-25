package main

import (
	win "calendar-scrapper/internal/winlosetie"
)

func main() {
	win.InitCmd()
	if err := win.Cmd.Execute(); err != nil {
		panic(err)
	}
}
