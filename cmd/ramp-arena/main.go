package main

import (
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var url = "https://api3.rampinteractive.com/livebarn/arenas/updatelist"
var client = &http.Client{}

type Result struct {
	RARID     uint   `json:"RARID"`
	SurfaceID string `json:"liveBarnId"`
}

func main() {
	config.Init("config", ".")
	cfg := config.MustReadConfig()
	repo := repository.NewRepository(cfg)
	db := repo.DB

	sql := "select RARID, surface_id FROM RAMP_Locations where surface_id<>0"

	var result []Result

	if err := db.Raw(sql).Scan(&result).Error; err != nil {
		panic(err)
	}

	if len(result) == 0 {
		log.Println("no data to send")
		return
	}

	token := login()

	data := map[string][]Result{
		"data": result,
	}

	s, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	sendData(string(s), token)
}

func sendData(data, token string) {
	url := "https://api3.rampinteractive.com/livebarn/arenas/updatelist"

	log.Println("sending: ", data)

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println("response: ", string(b))
}

func login() string {
	url := "https://api3.rampinteractive.com/auth/livebarn/authenticate"
	var data = strings.NewReader(
		`{"username":"",
    "password":"",
    "clientid" : "LiveBarn",
    "clientsecret": "F6CB1120-DF45-4621-BED3-0BDDE790B78F",
    "clientToken" : "LiveBarn"}`)

	req, err := http.NewRequest("POST", url, data)
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic("login resp not 200")
	}

	s, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(s))

	rdata := map[string]string{}
	err = json.Unmarshal(s, &rdata)
	if err != nil {
		panic(err)
	}

	return rdata["access_token"]
}
