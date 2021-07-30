package steamkeygrep

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"steamkeygrep/internal/logger"
	"strconv"
	"time"

	"github.com/valyala/fastjson"
)

type Pr0JsonInfo struct {
	start       bool
	end         bool
	lastPromote string
}

type MessageWithKey struct{}

func GrepPr0gramm(url string, tinMs time.Duration, keys chan<- string, errs chan<- error, stop <-chan bool, token string) {
	js := make(chan []byte, 10)
	contentJs := make(chan []byte, 10)
	comments := make(chan string)
	//contentIds := make(chan string)
	urls := make(chan string)
	commentUrls := make(chan string)
	pr0JsonInfos := make(chan Pr0JsonInfo)
	re, err := regexp.Compile("([0-9A-Z]{5}-[0-9A-Z]{5}-[0-9A-Z]{5})")
	if err != nil {
		errs <- fmt.Errorf("Can not compile regex\n%v", err)
		return
	}

	go func() {
		defer close(js)
		getJsonFromUrl(urls, js, errs, token)
	}()

	go func() {
		defer close(commentUrls)
		defer close(pr0JsonInfos)
		buildContentURLFromJSON(js, commentUrls, pr0JsonInfos, errs)
	}()

	go func() {
		defer close(contentJs)
		getJsonFromUrl(commentUrls, contentJs, errs, token)
	}()

	go func() {
		defer close(comments)
		grepCommentFromJSON(contentJs, comments, errs)
	}()

	go func() {
		getKeyFromComment(comments, keys, re)
	}()

	defer close(urls)
	urls <- url
	for {
		select {
		case pji := <-pr0JsonInfos:
			updatedUrl := url + "&older=" + pji.lastPromote
			urls <- updatedUrl
		case <-stop:
			return
		}
	}
}

func getJsonFromUrl(urls <-chan string, js chan<- []byte, errs chan<- error, token string) {
	for u := range urls {
		client := &http.Client{}
		defer client.CloseIdleConnections()
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			errs <- err
		}
		req.AddCookie(&http.Cookie{Name: "me", Value: token})
		resp, err := client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("error on making request to receive json\n%v", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errs <- fmt.Errorf("can not read response body\n%v", err)
		}
		resp.Body.Close()
		fmt.Println(string(body)) // TODO: error handling
		js <- body
	}
}

func buildContentURLFromJSON(js <-chan []byte, contentUrl chan<- string, pjis chan<- Pr0JsonInfo, errs chan<- error) {
	var pp fastjson.ParserPool
	for data := range js {
		p := pp.Get()
		defer pp.Put(p)
		v, err := p.ParseBytes(data)
		if err != nil {
			errs <- err
			continue
		}
		var lastPromote string
		var start, end bool
		for _, child := range v.GetArray("items") {
			//fmt.Printf("Comment %v:\n%v\n", i, string(child.GetStringBytes(contentKey)))
			id := strconv.Itoa(child.GetInt("id"))
			contentUrl <- "https://pr0gramm.com/api/items/info/get?itemId=" + id
			lastPromote = strconv.Itoa(child.GetInt("promoted"))
			start = child.GetBool("atStart")
			end = child.GetBool("atEnd")
		}
		if pjis != nil {
			pjis <- Pr0JsonInfo{
				start:       start,
				end:         end,
				lastPromote: lastPromote,
			}
		}

	}
}

func grepCommentFromJSON(js <-chan []byte, comments chan<- string, errs chan<- error) {
	var pp fastjson.ParserPool
	for data := range js {
		p := pp.Get()
		defer pp.Put(p)
		v, err := p.ParseBytes(data)
		if err != nil {
			errs <- err
			continue
		}
		for _, child := range v.GetArray("comments") {
			//fmt.Printf("Comment %v:\n%v\n", i, string(child.GetStringBytes(contentKey)))
			comments <- string(child.GetStringBytes("content"))
		}
	}
}

func getKeyFromComment(comments <-chan string, keys chan<- string, re *regexp.Regexp) {
	for c := range comments {
		fs := re.FindAllString(c, -1)
		if fs != nil {
			logger.Log.Infof("found keys in comment: %s\n", c)
			for _, key := range fs {
				keys <- key
			}
		}
	}
}
