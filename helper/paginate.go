package helper

import (
	"fmt"
	"math"
	"registration-service/dto"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Pagination(ctx *gin.Context) dto.PaginationRequest {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))

	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if err != nil {
		limit = 10
	}
	offset := (page - 1) * limit

	url := ctx.Request.URL
	url.RawQuery = ""

	if url.Path[len(url.Path)-1] == '/' {
		url.Path = url.Path[:len(url.Path)-1]
	}

	return dto.PaginationRequest{
		Offset: offset,
		Limit:  limit,
		URL:    url.String(),
	}

}

func TotalPages(limit int, total_data int64) int64 {
	return (total_data + int64(limit) - 1) / int64(limit)
}

func GeneratePageURL(baseURL string, page, limit int) string {
	if page < 1 {
		return ""
	}
	return fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page, limit)
}

func MetaDataPagination(total_data int64, pagReq dto.PaginationRequest) dto.PaginationResponse {

	total_pages := TotalPages(pagReq.Limit, total_data)

	currentPage := (pagReq.Offset / pagReq.Limit) + 1

	return dto.PaginationResponse{
		CurrentPage:  currentPage,
		FirstPageUrl: GeneratePageURL(pagReq.URL, 1, pagReq.Limit),
		LastPage:     total_pages,
		LastPageUrl:  GeneratePageURL(pagReq.URL, int(total_pages), pagReq.Limit),
		NextPageUrl:  GeneratePageURL(pagReq.URL, currentPage+1, pagReq.Limit),
		PerPage:      pagReq.Limit,
		PrevPageUrl:  GeneratePageURL(pagReq.URL, currentPage-1, pagReq.Limit),
		To:           int(math.Min(float64(pagReq.Offset+pagReq.Limit), float64(total_data))),
		Total:        total_data,
		TotalPages:   total_pages,
	}
}
