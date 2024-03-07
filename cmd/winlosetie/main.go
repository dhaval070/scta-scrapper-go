package main

import (
	"calendar-scrapper/config"
	win "calendar-scrapper/internal/winlosetie"
	"calendar-scrapper/pkg/repository"
)

func main() {
	config.Init("config", ".")
	cfg := config.MustReadConfig()
	repo := repository.NewRepository(cfg)
	win.InitCmd(repo)
	win.Cmd.Execute()
}
