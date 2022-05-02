package service

import "muti-kube/models/common"

type OpOption func(*Op)

type Op struct {
	pagination *common.Pagination
}

func WithPagination(pagination *common.Pagination) OpOption {
	return func(op *Op) { op.pagination = pagination }
}
