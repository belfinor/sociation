package sociation

// @author  Mikhail Kirillov <mikkirillov@yandex.ru>
// @version 1.000
// @date    2018-06-26

import (
	"context"
	"testing"
)

func TestGet(t *testing.T) {
	list := GetWords(context.Background(), "игра")
	if list == nil || len(list) == 0 {
		t.Fatal("client failed")
	}
}
