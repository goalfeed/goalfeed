package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func GetString(url string, ret chan string) {
	var bodyChan chan []byte = make(chan []byte)
	go GetByte(url, bodyChan)
	bodyBytes := <-bodyChan
	ret <- string(bodyBytes)
}

func GetByte(url string, ret chan []byte) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		fmt.Printf("%+v\n", err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	ret <- bodyBytes
}
