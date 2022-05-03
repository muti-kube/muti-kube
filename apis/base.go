package apis

import (
	"muti-kube/models/common"
	"muti-kube/pkg/util/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Base struct{}

func (b *Base) GetPagination(c *gin.Context) *common.Pagination {
	//page and size query parameter handle
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "500"))
	order := c.DefaultQuery("order", "desc")
	return &common.Pagination{
		PageSize: pageSize,
		Page:     page,
		Order:    order,
	}
}

func (b *Base) OK(c *gin.Context, data interface{}, msg string) {
	var res common.Response
	res.Data = data
	if msg != "" {
		res.Msg = msg
	}
	c.JSON(http.StatusOK, res.ReturnOK())
}

func (b *Base) PageOK(c *gin.Context, result interface{}, count *int64, page *common.Pagination, msg string) {
	var res common.PageResponse
	res.Data.List = result
	res.Data.Count = count
	res.Data.PageIndex = page.Page
	res.Data.PageSize = page.PageSize
	if msg != "" {
		res.Msg = msg
	}
	c.JSON(http.StatusOK, res.ReturnOK())
}

func (b *Base)Error(c *gin.Context, code int, err error, msg string) {
	var res common.Response
	res.Msg = err.Error()
	if msg != "" {
		res.Msg = msg
	}
	logger.Error(res.Msg)
	c.JSON(http.StatusOK, res.ReturnError(code))
}