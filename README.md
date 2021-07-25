# 评论系统优化

原项目中app与网页等不同主题内容皆需要评论，现参照`评论系统设计`将评论系统独立出来，作为公共的服务
> 项目结构参照 [beer-shop](https://github.com/go-kratos/beer-shop) 目录结构，基于 [kratos](https://github.com/go-kratos/kratos) 框架搭建

### 1. 项目模块
- account-service 账户服务
- comment-service 评论服务
- comment-job  评论消息消费任务
- baseapp-interface BFF层

### 2. 依赖组件
- consul 服务注册发现
- kafka 消息队列
- redis 缓存
- gorm orm框架

### 3. 并行请求示例
```go
// 查询评论信息后，查询相关的用户信息以及评论点赞信息
func (uc *CommentUsecase) GetCommentList(ctx context.Context, subject *CommentSubject,
	page, size, replyCount int) ([]*Comment, error) {
	comments, err := uc.repo.GetCommentList(ctx, subject, page, size, replyCount)
	if err != nil {
		return nil, err
	}
	// 并发查询附加信息
	group, _ := errgroup.WithContext(ctx)
	var accounts []*Account
	likedMap := make(map[uint64]bool)
	group.Go(func() error {
		accounts, err = uc.accountRepo.ListByIds(ctx, getCommentMemberIds(comments))
		return err
	})
	// 如果有用户信息，查询用户是否点过赞
	if uid, err := token.ExtractUid(ctx); err == nil {
		group.Go(func() error {
			temp, err := uc.repo.GetLikeItem(ctx, uid, getCommentIds(comments))
			likedMap = temp
			return err
		})
	}
	if err = group.Wait(); err != nil {
		return nil, err
	}
	concatMemberInfo(comments, accounts, likedMap)
	return comments, nil
}
```

### 4. 错误码使用
接口使用protobuf定义错误，并在接口返回时使用，系统内部使用 kratos `errors.new(code, reason, message)`方法包装错误，依据code进行判定。
数据库错误将被包装，避免orm框架的错误传播，产生强依赖。

