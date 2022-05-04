package common

type Pagination struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Order    string `json:"order" form:"order"`
}

type Page struct {
	List      interface{} `json:"list"`
	Count     *int64      `json:"count"`
	PageIndex int         `json:"page_size"`
	PageSize  int         `json:"page_size"`
}

type Response struct {
	Code int         `json:"code" example:"200"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

type PageResponse struct {
	Code int    `json:"code" example:"200"`
	Data Page   `json:"data"`
	Msg  string `json:"msg"`
}

func (res *Response) ReturnOK() *Response {
	res.Code = 200
	return res
}

func (res *Response) ReturnError(code int) *Response {
	res.Code = code
	return res
}

func (res *PageResponse) ReturnOK() *PageResponse {
	res.Code = 200
	return res
}
