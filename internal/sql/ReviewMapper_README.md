# ReviewMapper 使用文档

## 概述
ReviewMapper 提供了对评论（Review）数据的完整 CRUD 操作，支持多种查询方式和批量操作。

## 初始化
```go
import "blog/internal/sql"

db, _ := sqlx.Connect("mysql", "user:password@tcp(localhost:3306)/database")
reviewMapper := sql.NewReviewMapper(db)
```

## CRUD 操作

### Create - 创建评论

#### 单条创建
```go
newReview := &model.Review{
    ArticleID: 1,
    Content:    "这是一篇很棒的文章！",
    AuthorID:   100,
    IsDirect:   true,
    ParentID:   0,
}

id, err := reviewMapper.Create(newReview)
if err != nil {
    log.Printf("创建评论失败: %v", err)
}
fmt.Printf("评论ID: %d\n", id)
```

#### 批量创建
```go
reviews := []*model.Review{
    {ArticleID: 1, Content: "评论1", AuthorID: 100, IsDirect: true, ParentID: 0},
    {ArticleID: 1, Content: "评论2", AuthorID: 100, IsDirect: true, ParentID: 0},
}

lastID, err := reviewMapper.BatchCreate(reviews)
```

### Read - 查询评论

#### 按ID查询
```go
review, err := reviewMapper.FindByID(1)
```

#### 按文章ID查询
```go
reviews, err := reviewMapper.FindByArticleID(1)
```

#### 按作者ID查询
```go
reviews, err := reviewMapper.FindByAuthorID(100)
```

#### 按父评论ID查询（获取回复）
```go
replies, err := reviewMapper.FindByParentID(1)
```

#### 查询直接评论
```go
directReviews, err := reviewMapper.FindDirectReviews(1)
```

#### 查询回复
```go
replies, err := reviewMapper.FindReplies(1)
```

#### 分页查询
```go
reviews, err := reviewMapper.FindByArticleIDWithPagination(1, 10, 0)
```

#### 时间范围查询
```go
startTime := time.Now().AddDate(0, -1, 0)
endTime := time.Now()
reviews, err := reviewMapper.FindWithTimeRange(1, startTime, endTime)
```

#### 查询最新评论
```go
reviews, err := reviewMapper.FindLatestByAuthorID(100, 5)
```

#### 查询最近评论
```go
reviews, err := reviewMapper.FindRecentReviews(10)
```

#### 查询热门评论
```go
reviews, err := reviewMapper.FindPopularReviews(10)
```

### Update - 更新评论

#### 更新整个评论
```go
review, _ := reviewMapper.FindByID(1)
review.Content = "更新后的内容"
err := reviewMapper.Update(review)
```

#### 只更新内容
```go
err := reviewMapper.UpdateContent(1, "新的评论内容")
```

### Delete - 删除评论

#### 按ID删除
```go
err := reviewMapper.DeleteByID(1)
```

#### 按文章ID删除（删除文章的所有评论）
```go
err := reviewMapper.DeleteByArticleID(1)
```

#### 按作者ID删除（删除用户的所有评论）
```go
err := reviewMapper.DeleteByAuthorID(100)
```

#### 按父评论ID删除（删除评论的所有回复）
```go
err := reviewMapper.DeleteByParentID(1)
```

## 辅助查询方法

### 统计查询
```go
// 总评论数
total, _ := reviewMapper.FindAllCount()

// 文章评论数
count, _ := reviewMapper.FindByArticleIDCount(1)

// 用户评论数
count, _ := reviewMapper.CountByAuthorID(100)

// 回复数量
count, _ := reviewMapper.CountReplies(1)
```

### 存在性检查
```go
exists, err := reviewMapper.ExistsByID(1)
```

## 方法列表

### 创建操作
- `Create(review *model.Review) (int64, error)` - 创建单条评论
- `BatchCreate(reviews []*model.Review) (int64, error)` - 批量创建评论

### 查询操作
- `FindByID(id int) (*model.Review, error)` - 按ID查询
- `FindByArticleID(articleID int) ([]*model.Review, error)` - 按文章ID查询
- `FindByAuthorID(authorID int) ([]*model.Review, error)` - 按作者ID查询
- `FindByParentID(parentID int) ([]*model.Review, error)` - 按父评论ID查询
- `FindDirectReviews(articleID int) ([]*model.Review, error)` - 查询直接评论
- `FindReplies(parentID int) ([]*model.Review, error)` - 查询回复
- `FindAll(limit int, offset int) ([]*model.Review, error)` - 查询所有评论（分页）
- `FindAllCount() (int, error)` - 查询总评论数
- `FindByArticleIDWithPagination(articleID int, limit int, offset int) ([]*model.Review, error)` - 分页查询文章评论
- `FindByArticleIDCount(articleID int) (int, error)` - 查询文章评论数
- `FindLatestByAuthorID(authorID int, limit int) ([]*model.Review, error)` - 查询用户最新评论
- `FindByArticleIDAndAuthorID(articleID int, authorID int) ([]*model.Review, error)` - 查询用户在某文章的评论
- `FindWithTimeRange(articleID int, startTime time.Time, endTime time.Time) ([]*model.Review, error)` - 时间范围查询
- `FindRecentReviews(limit int) ([]*model.Review, error)` - 查询最近评论
- `FindPopularReviews(limit int) ([]*model.Review, error)` - 查询热门评论

### 更新操作
- `Update(review *model.Review) error` - 更新整个评论
- `UpdateContent(id int, content string) error` - 只更新内容

### 删除操作
- `DeleteByID(id int) error` - 按ID删除
- `DeleteByArticleID(articleID int) error` - 按文章ID删除
- `DeleteByAuthorID(authorID int) error` - 按作者ID删除
- `DeleteByParentID(parentID int) error` - 按父评论ID删除

### 辅助方法
- `ExistsByID(id int) (bool, error)` - 检查评论是否存在
- `CountByArticleID(articleID int) (int, error)` - 统计文章评论数
- `CountByAuthorID(authorID int) (int, error)` - 统计用户评论数
- `CountReplies(parentID int) (int, error)` - 统计回复数

## 数据模型
```go
type Review struct {
    ID         int       `db:"id"`          // 评论ID
    ArticleID  int       `db:"article_id"`  // 评论文章ID
    CreateTime time.Time `db:"create_time"` // 创建时间
    UpdateTime time.Time `db:"update_time"` // 更新时间
    Content    string    `db:"content"`     // 评论内容
    AuthorID   int       `db:"author_id"`   // 评论作者ID
    IsDirect   bool      `db:"is_direct"`   // 是否直接评论
    ParentID   int       `db:"parent_id"`   // 父评论ID
}
```

## 特性

1. **完整的 CRUD 操作** - 支持创建、读取、更新、删除
2. **多种查询方式** - 支持按不同条件查询
3. **分页支持** - 支持分页查询
4. **批量操作** - 支持批量创建
5. **时间范围查询** - 支持按时间范围查询
6. **热门评论** - 支持查询热门评论（按回复数排序）
7. **错误处理** - 所有方法都包含错误处理
8. **参数绑定** - 使用 sqlx 的参数绑定，防止 SQL 注入
9. **结果映射** - 自动映射到结构体

## 注意事项

1. 所有查询方法都使用参数绑定，防止 SQL 注入
2. 时间字段自动设置为当前时间
3. 批量操作使用事务，确保数据一致性
4. 删除操作不会级联删除，需要手动处理
5. 热门评论按回复数和创建时间排序