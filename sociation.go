package sociation

// @author  Mikhail Kirillov <mikkirillov@yandex.ru>
// @version 1.005
// @date    2019-02-19

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

var cache *lcache.Cache = lcache.New(&lcache.Config{TTL: 86400 * 30, Size: 5000, Nodes: 24})

var SERVER_URL string = "https://sociation.org/ajax/word_associations/"

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

	res, err := ua.Request("POST", SERVER_URL, headers, []byte(content))
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

		out := make(chan *SociumResp, 1)

		go func() {

			res := fetch(word)
			if res != nil {
				out <- res
			}
			close(out)

		}()

		var res []Result

		select {
		case resp, ok := <-out:
			if ok {
				res = resp.Associations
			}
		case <-ctx.Done():
		}

		return res
	})

	if r == nil {
		return []Result{}
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
