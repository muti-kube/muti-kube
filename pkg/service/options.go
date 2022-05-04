package service

import "muti-kube/models/common"

type OpOption func(*Op)

type Op struct {
	Pagination *common.Pagination
}

func WithPagination(pagination *common.Pagination) OpOption {
	return func(op *Op) { op.Pagination = pagination }
}

func (op *Op) applyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(op)
	}
}

func OpGet(opts ...OpOption) Op {
	ret := Op{}
	ret.applyOpts(opts)
	return ret
}
