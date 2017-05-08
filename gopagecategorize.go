package gopagecategorize

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/PuerkitoBio/goquery"
)

var score = map[string]float32{
	"title": 0.15,
	"h1":    0.13,
	"h2":    0.11,
	"h3":    0.07,
	"h4":    0.05,
	"h5":    0.03,
	"h6":    0.02,
	"em":    0.01,
	"b":     0.01,
	"i":     0.01,
	"ins":   0.01,
	"s":     0.01,
	"a":     0.005,
	"del":   -0.01,
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.1 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.1",
	"Mozilla/5.0 (Windows NT 6.3; rv:36.0) Gecko/20100101 Firefox/36.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko",
	"Mozilla/5.0 (compatible, MSIE 11, Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A",
	"Mozilla/5.0 (iPad; CPU OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5355d Safari/8536.25",
	"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	"Googlebot/2.1 (+http://www.googlebot.com/bot.html)",
	"Googlebot/2.1 (+http://www.google.com/bot.html)",
	"Googlebot-Image/1.0",
}
var timeout = time.Second * 100

type hitResponse struct {
	body     string
	rawBody  []byte
	url      string
	duration int //in nano seconds
}
type hitRequest struct {
	url          string
	timeout      time.Duration
	userAgent    string
	params       string
	method       string
	needResponse bool
	referer      string
}
type scoreBoardStruct struct {
	tag   string
	score float32
}
type ByScore []scoreBoardStruct

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].score > a[j].score }

func AnalyzeUrl(url string) ([]scoreBoardStruct, error) {

	var scoreBoard = make(map[string]float32)
	var scoreSortableBoard []scoreBoardStruct
	userAgent := fun.Sample(userAgents, 1).([]string)[0]
	//lets send custom UA
	ht := hitRequest{
		timeout:      timeout,
		userAgent:    userAgent,
		url:          url,
		needResponse: true,
	}
	hr, _ := hit(ht)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(hr.body))
	if err != nil {
		log.Fatal(err)
		return scoreSortableBoard, err
	}
	var text string
	for tag, _ := range score {
		doc.Find(tag).Each(func(i int, s *goquery.Selection) {
			text = s.Text()

			if text == "" {
				return
			}
			allWords := strings.Split(text, " ")
			for _, oneWord := range allWords {
				oneWord = strings.Trim(oneWord, " \n")
				if oneWord == "" {
					continue
				}
				scoreBoard[oneWord] += score[tag]
			}
		})
	}
	for sb_key, sb_value := range scoreBoard {
		scoreSortableBoard = append(scoreSortableBoard, scoreBoardStruct{sb_key, sb_value})
	}
	sort.Sort(ByScore(scoreSortableBoard))
	return scoreSortableBoard, nil
}

//Hits a page and return body if needed
func hit(ht hitRequest) (hitResponse, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	hr := hitResponse{}
	hr.url = ht.url
	req, err := http.NewRequest(ht.method, ht.url, nil)
	if err != nil {
		return hr, errors.New("Can't hit it")
	}
	if ht.userAgent != "" {
		req.Header.Set("User-Agent", ht.userAgent)
	}

	if ht.referer != "" {
		req.Header.Set("Referer", ht.referer)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return hr, errors.New("Can't hit it")
	}

	defer resp.Body.Close()
	if ht.needResponse == false {
		return hr, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return hr, errors.New("Can't hit it")
	}
	hr.body = string(body)
	hr.rawBody = body
	return hr, nil
}
