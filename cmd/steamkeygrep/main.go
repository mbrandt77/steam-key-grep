package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"steamkeygrep/internal/logger"
	"steamkeygrep/pkg/steamkeygrep"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("conf.yaml")
	viper.AddConfigPath("/config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Log.Fatalf("fatal error on reading config file: %v\n", err)
	}

	//newOnlyUrlAll := "https://pr0gramm.com/api/items/get?flags=15"
	//promotedUrlAll := "https://pr0gramm.com/api/items/get?flags=15&promoted=1"
	url := "https://pr0gramm.com/api/items/get?flags=6&promoted=1"
	tinMs := 300 * time.Millisecond
	keys := make(chan string)
	errs := make(chan error)
	restart := make(chan bool)

	logger.Log.Infow("Start logging", "url", url)

	go sendKeyToDiscord(viper.GetString("WEBHOOK_URL"), keys, errs)

	doGrep(url, tinMs, keys, errs, restart, viper.GetString("REQUEST_PR0_TOKEN"))
}

func doGrep(url string, tinMs time.Duration, keys chan<- string, errs chan<- error, restart <-chan bool, token string) {
	stop := make(chan bool)
	go steamkeygrep.GrepPr0gramm(url, tinMs, keys, errs, stop, token)
	for {
		select {
		case <-restart:
			stop <- true
			go steamkeygrep.GrepPr0gramm(url, tinMs, keys, errs, stop, token)
		}
	}
}

func sendKeyToDiscord(wu string, keys <-chan string, errs chan<- error) {
	for key := range keys {
		reqBody, err := json.Marshal(map[string]string{
			"content": key,
		})
		if err != nil {
			errs <- errors.Wrap(err, "error on creating request body")
		}

		req, err := http.NewRequest("POST", wu, bytes.NewBuffer(reqBody))
		if err != nil {
			errs <- errors.Wrap(err, "can not create request")
		}
		req.Header.Set("Content-Type", "application/json")

		client := http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			errs <- errors.Wrap(err, "can execute request")
		}
		if resp.StatusCode >= 300 {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errs <- errors.Wrap(err, "can not read respone body")
			}
			logger.Log.Warnf("response from statuscode: %v\nresp: %v\n", resp.StatusCode, string(b))
			resp.Body.Close()
		}
	}
}

func printErrors(errs <-chan error) {
	for err := range errs {
		logger.Log.Error(err)
	}
}
