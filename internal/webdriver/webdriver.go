package webdriver

import (
	"log"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func GetWebDriver() selenium.WebDriver {
	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"--headless", // comment out this line for testing
		},
		Path: "/home/dhaval/nodejs/chr/chrome/linux-126.0.6478.126/chrome-linux64/chrome",
	})

	// create a new remote client with the specified options
	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		log.Fatal("new remote", err)
	}
	return driver
}
