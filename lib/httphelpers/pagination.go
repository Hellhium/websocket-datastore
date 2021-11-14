package httphelpers

import (
	"net/url"
	"strconv"
)

func Pagination(url *url.URL) (int, int) {

	iSkip := 0
	skip := url.Query().Get("skip")
	if skip != "" {
		iSkip, _ = strconv.Atoi(skip)
	}
	iLimit := 20
	limit := url.Query().Get("limit")
	if limit != "" {
		iLimit, _ = strconv.Atoi(limit)
	}

	if iSkip < 0 {
		iSkip = 0
	}
	if iLimit < 1 {
		iLimit = 1
	}
	if iLimit > 150 {
		iLimit = 150
	}

	return iLimit, iSkip
}
