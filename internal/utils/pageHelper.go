package utils

// 分页助手
//
//	@param	page		页码
//	@param	pageSize	每页条数
//	@return	page 页码
//	@return	pageSize 每页条数
const (
	DefaultPageSize = 10
	MaxPageSize     = 100
	// MaxOffset 用于限制深分页导致的慢查询风险，可按业务调整
	MaxOffset = 10000
)

func GetPage(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}
func GetOffset(page int, pageSize int) int {
	return (page - 1) * pageSize
}
func GetLimit(pageSize int) int {
	return pageSize
}
func GetTotalPage(total int, pageSize int) int {
	return (total + pageSize - 1) / pageSize
}

// ResolveOffsetLimit 通过 page/pageSize 计算 offset/limit，并处理深分页问题
// 返回值:
// - offset: 可直接用于 SQL OFFSET
// - limit:  可直接用于 SQL LIMIT
func ResolveOffsetLimit(page int, pageSize int) (offset int, limit int) {
	page, pageSize = GetPage(page, pageSize)
	offset = GetOffset(page, pageSize)
	limit = GetLimit(pageSize)

	// 深分页保护：超过阈值时，固定在最大允许 offset
	if offset > MaxOffset {
		offset = MaxOffset
	}
	return offset, limit
}
