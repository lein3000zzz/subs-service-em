package utils

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetPageAndLimitFromContext(c *gin.Context) (int, int) {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 1000 {
			limit = val
		}
	}

	return page, limit
}

func CountPages(total, limit int64) int64 {
	pages := int64(1)
	if total > 0 {
		pages = total / limit
		if total%limit != 0 {
			pages++
		}
	}
	return pages
}

func GetOverlappedMonths(filterStart, filterEnd, actualStart time.Time, actualEnd *time.Time) int {
	start := maxDate(filterStart, actualStart)
	end := filterEnd

	if actualEnd != nil {
		subEnd := *actualEnd
		if subEnd.Before(end) {
			end = subEnd
		}
	}

	if end.Before(start) {
		return 0
	}

	y1, m1, _ := start.Date()
	y2, m2, _ := end.Date()
	return (y2-y1)*12 + int(m2-m1) + 1
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
