package util


func ConvertToInt64Ptr(x int) *int64 {
	xInt64 := int64(x)
	return &xInt64
}
