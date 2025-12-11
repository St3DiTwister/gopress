package httpx

import (
	"net/url"
	"strconv"
)

func QueryInt(q url.Values, name string, def int) int {
	s := q.Get(name)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
