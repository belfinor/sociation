package sociation

// @author  Mikhail Kirillov <mikkirillov@yandex.ru>
// @version 1.001
// @date    2018-06-27

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/belfinor/Helium/log"
	"github.com/belfinor/Helium/net/http/client"
	"github.com/belfinor/lcache"
)

type Result struct {
	Name    string `json:"name"`
	Direct  int64  `json:"popularity_direct"`
	Inverse int64  `json:"popularity_inverse"`
}

type SociumResp struct {
	Associations []Result `json:"associations"`
	Word         string   `json:"word"`
}

type Request struct {
	Word string
	Out  chan *SociumResp
}

var cache *lcache.Cache = lcache.New(&lcache.Config{TTL: 86400, Size: 5000, Clean: 100, InputBuffer: 100, Nodes: 24})
var input chan *Request = make(chan *Request, 100)

func worker() {

	for r := range input {

		res := fetch(r.Word)
		if res != nil {
			r.Out <- res
		}
		close(r.Out)
	}
}

func init() {
	go worker()
}

func fetch(phrase string) *SociumResp {

	form := url.Values{}

	form.Add("max_count", "0")
	form.Add("back", "false")
	form.Add("word", phrase)

	content := form.Encode()

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer":      "http://sociation.org/graph/",
		"User-Agent":   "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36",
	}

	ua := client.New()
	ua.Timeout = time.Second * 10

	res, err := ua.Request("POST", "http://sociation.org/ajax/word_associations/", headers, []byte(content))
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	resp := new(SociumResp)
	if err = json.Unmarshal(res.Content, &resp); err != nil {
		log.Error(err.Error())
		return nil
	}

	return resp
}

func Get(ctx context.Context, word string) []Result {

	r := cache.Fetch(word, func(key string) interface{} {
		req := &Request{
			Word: word,
			Out:  make(chan *SociumResp, 1),
		}

		input <- req

		var res []Result

		select {
		case resp, ok := <-req.Out:
			if ok {
				res = resp.Associations
			}
		case <-ctx.Done():
		}

		return res
	})

	if r == nil {
		return nil
	}

	return r.([]Result)
}

func GetWords(ctx context.Context, word string) []string {
	res := Get(ctx, word)
	if res == nil || len(res) == 0 {
		return nil
	}

	result := make([]string, len(res))

	for i, v := range res {
		result[i] = v.Name
	}

	return result
}
