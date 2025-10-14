package utils

import (
	"strconv"

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
