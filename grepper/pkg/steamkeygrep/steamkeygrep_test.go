package steamkeygrep

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

func TestGetKeyFromComment(t *testing.T) {
	commentChan := make(chan string)
	keys := make(chan string)
	expectedFirstKey := "6RTHD-GR5TB-H4LOP"
	expectedSecondKey := "H4LOP-GR5TB-6RTHD"
	re, err := regexp.Compile("([0-9A-Z]{5}-[0-9A-Z]{5}-[0-9A-Z]{5})")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		commentChan <- "hier ich habe einen Key für euch: 6RTHD-GR5TB-H4LOP  Viel Spaß damit"
		commentChan <- "jojo leute ich habe nen key für euch H4LOP-GR5TB-6RTHD"
	}()
	go getKeyFromComment(commentChan, keys, re)
	select {
	case key := <-keys:
		if key != expectedFirstKey {
			t.Errorf("Received unexpected key!\nexpected: %v\nactual: %v", expectedFirstKey, key)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Not reveived Key after 2 Seconds")
	}

	select {
	case key := <-keys:
		if key != expectedSecondKey {
			t.Errorf("Received unexpected key!\nexpected: %v\nactual: %v", expectedSecondKey, key)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Not reveived Key after 2 Seconds")
	}
}

func TestGetMultipleKeysFromComment(t *testing.T) {
	commentChan := make(chan string)
	keys := make(chan string)
	expectedAmountKeys := 15
	foundKeys := make([]string, 0, 0)
	re, err := regexp.Compile("([0-9A-Z]{5}-[0-9A-Z]{5}-[0-9A-Z]{5}-[0-9A-Z]{5}-[0-9A-Z]{5}|[0-9A-Z]{5}-[0-9A-Z]{5}-[0-9A-Z]{5})")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer close(commentChan)
		commentChan <- "Crown Champion Legends of the Arena: W7R8H-NL6JL-EGT5W | The 39 Steps: PCDM9-KRQD8-ER2P9 | Freedom Force vs. the Third Reich : G2MWX-HE440-RXBNG | Showtime: N0L53-XLNK9-M5TQZ | Cobi Treasure Deluxe: YAIDN-0G6XA-J3DMI | Ionball 2 Ionstorm: ZLQ8V-6J752-ERWYC | Zombie Ballz: GJX0A-PZ3V8-VW6J3 | Jack Orlando Director's Cut: FQ3RE-QZ0X4-LY5VC | One Hundred Ways: NLH3V-4TE6J-FH68C | Puzzle Chambers: 6LMDQ-Y353F-7AX9Z"
		commentChan <- "OT:\r\nDFCC8-3ELFE-LERPM  |  The Bard's Tale IV: Director's Cut \r\nD9Q7H-DBYLP-J2VLF  |  Shoppe Keep 2 \r\nCB5PP-DYWPV-NZ9X7  |  Capitalism 2 \r\nEX83A-N9P7T-N0495  |  Raiden V: Director's Cut | 雷電 V Director's Cut | 雷電V:導演剪輯版 \r\nFXWEP-IMDFI-EIZD2  |  Truberbrook / Trüberbrook \r\n\r\nJa nehmen, gn8.\n"
	}()
	go func() {
		defer close(keys)
		getKeyFromComment(commentChan, keys, re)
	}()

	for key := range keys {
		foundKeys = append(foundKeys, key)
	}

	if len(foundKeys) != expectedAmountKeys {
		t.Errorf("amount of keys is not as expected\nexpected: %v\nactual: %v\n", expectedAmountKeys, len(foundKeys))
	}
}

func TestBuildOneContentURLFromJSON(t *testing.T) {
	errs := make(chan error)
	jschan := make(chan []byte)
	contentUrl := make(chan string)
	//pjis := make(chan Pr0JsonInfo)
	expectedUrl := "https://pr0gramm.com/api/items/info/get?itemId=3359331"
	data, err := ioutil.ReadFile("testdata/testBuildOneContentURLFromJSON.json")
	if err != nil {
		t.Fatalf("can not read file!\n%v", err)
	}

	go func() {
		jschan <- data
	}()

	go buildContentURLFromJSON(jschan, contentUrl, nil, errs)

	select {
	case cu := <-contentUrl:
		if cu != expectedUrl {
			t.Errorf("received non expected url\nexpected: %v\nactual: %v\n", expectedUrl, cu)
		}
	case <-time.After(3 * time.Second):
		t.Errorf("not received url after 3 seconds")
	}
}

func getGoldenDataMap(goldenFilePath string) (map[string]bool, error) {
	goldendata, err := ioutil.ReadFile(goldenFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "can not read golden testfile")
	}

	var goldenDataJson []map[string]interface{}
	err = json.Unmarshal(goldendata, &goldenDataJson)
	if err != nil {
		return nil, errors.Wrap(err, "can not unmarshal goldendata")
	}

	goldenDataMap := make(map[string]bool)

	for _, data := range goldenDataJson {
		goldenDataMap[data["id"].(string)] = true
	}

	if len(goldenDataMap) != len(goldenDataJson) {
		return nil, errors.Wrap(err, "amount of goldenmap not equals goldenSlice")
	}

	return goldenDataMap, nil

}

func TestBuildMultipleContentURLFromOneJSON(t *testing.T) {
	errs := make(chan error)
	jschan := make(chan []byte)
	contentUrl := make(chan string)
	//pjis := make(chan Pr0JsonInfo)

	data, err := ioutil.ReadFile("testdata/testfulljson.json")
	if err != nil {
		t.Fatalf("can not read file!\n%v", err)
	}

	goldenDataMap, err := getGoldenDataMap("testdata/golden/goldentestfulljson.json")
	if err != nil {
		t.Fatalf("Can not create goldenDataMap\n%v", err)
	}

	go func() {
		jschan <- data
	}()

	go buildContentURLFromJSON(jschan, contentUrl, nil, errs)
	run := true
	for run {
		select {
		case cu := <-contentUrl:
			if ok := goldenDataMap[cu]; !ok {
				t.Errorf("received non expected url\nactual: %v\n", cu)
				continue
			}
			delete(goldenDataMap, cu)
		case <-time.After(2 * time.Second):
			run = false
		}
	}

	if len(goldenDataMap) > 0 {
		t.Errorf("There are Urls left\nLeft: %v", len(goldenDataMap))
		for url, _ := range goldenDataMap {
			fmt.Println(url)
		}
	}
}

func TestGetNextPage(t *testing.T) {
	errs := make(chan error)
	jschan := make(chan []byte)
	contentUrl := make(chan string)
	pjis := make(chan Pr0JsonInfo)
	expectedPromotedID := "454944"
	actualPromotedID := ""

	data, err := ioutil.ReadFile("testdata/testfulljson.json")
	if err != nil {
		t.Fatalf("can not read file!\n%v", err)
	}
	goldenDataMap, err := getGoldenDataMap("testdata/golden/goldentestfulljson.json")
	if err != nil {
		t.Fatalf("Can not create goldenDataMap\n%v", err)
	}

	go func() {
		jschan <- data
	}()

	go buildContentURLFromJSON(jschan, contentUrl, pjis, errs)
	run := true
	for run {
		select {
		case cu := <-contentUrl:
			if ok := goldenDataMap[cu]; !ok {
				t.Errorf("received non expected url\nactual: %v\n", cu)
				continue
			}
			delete(goldenDataMap, cu)
		case actualPr0JsonInfo := <-pjis:
			actualPromotedID = actualPr0JsonInfo.lastPromote
		case <-time.After(2 * time.Second):
			run = false
		}
	}

	if len(goldenDataMap) > 0 {
		t.Errorf("There are Urls left\nLeft: %v", len(goldenDataMap))
		for url, _ := range goldenDataMap {
			fmt.Println(url)
		}
	}

	if actualPromotedID != expectedPromotedID {
		t.Errorf("Received not expected promoted id\nexpected: %v\nactual: %v", expectedPromotedID, actualPromotedID)
	}

}

func TestGrepPr0gramm(t *testing.T) {
	t.Skip()
	viper.SetConfigFile("conf.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("fatal error on reading config file: %v\n", err)
	}
	url := "https://pr0gramm.com/api/items/get?flags=15&promoted=1"
	//url := "https://pr0gramm.com/api/items/get?flags=15"
	token := viper.GetString("REQUEST_PR0_TOKEN")
	errs := make(chan error)
	tinS := time.Duration(5000 * time.Millisecond)
	stop := make(chan bool, 1)
	keys := make(chan string)
	go GrepPr0gramm(url, tinS, keys, errs, stop, token)
	for {
		select {
		case err := <-errs:
			t.Errorf("Received unexpected error: %v", err)
		case k := <-keys:
			t.Logf("Received STEAMKEY: %v", k)
		case <-time.After(tinS * 10):
			t.Log("Received nothing")
			return
		}
	}
}

/*

//older
https://pr0gramm.com/api/items/get?flags=6&promoted=1&older=LASTPROMOTEID

//beliebt
//https://pr0gramm.com/api/items/get?flags=6&promoted=1

//neu
//https://pr0gramm.com/api/items/get?flags=6

//itemidbeispiel
//https://pr0gramm.com/api/items/info/get?itemId=idnr
*/
