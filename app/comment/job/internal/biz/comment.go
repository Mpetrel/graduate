package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"time"
)

// CommentSubject 评论主题对象
type CommentSubject struct {
	Id uint64
	ObjId uint64
	ObjType int
	MemberId uint64
	Count int
	RootCount int
	AllCount int
	State int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Comment 评论本身
type Comment struct {
	Id uint64
	MemberId uint64
	Root uint64
	Parent uint64
	ParentMemberId uint64 // 回复对象的作者id
	Floor int
	Count int
	RootCount int
	Like int
	Liked bool
	Hate int
	State int8
	AtMemberIds string
	Ip string
	Platform int8
	Device string
	Message string
	Meta string
	CreatedAt time.Time
	UpdatedAt time.Time
	Replies []*Comment
}


type LikeItem struct {
	CommentId uint64
	Like bool
}

type CommentIndexCache struct {
	ObjId uint64
	ObjType int
	Page int
	Size int
}

type CommentRepo interface {
	CreateSubject(ctx context.Context, subject *CommentSubject) error
	SaveComment(ctx context.Context, subject *CommentSubject, comment *Comment) error
	UpdateLikeNum(ctx context.Context, comment *Comment) error
	BuildCommentIndexCache(ctx context.Context, param CommentIndexCache) error
}

type CommentUsecase struct {
	repo CommentRepo
	log *log.Helper
}

func NewCommentUsecase(repo CommentRepo, logger log.Logger) *CommentUsecase {
	return &CommentUsecase{
		repo: repo,
		log: log.NewHelper(logger),
	}
}

func (uc *CommentUsecase) CreateSubject(ctx context.Context, subject *CommentSubject) error {
	return uc.repo.CreateSubject(ctx, subject)
}

// CreateComment 创建一条评论
func (uc *CommentUsecase) CreateComment(ctx context.Context, subject *CommentSubject, comment *Comment) error {
	return uc.repo.SaveComment(ctx, subject, comment)
}

func (uc *CommentUsecase) BuildCommentIndexCache(ctx context.Context, param CommentIndexCache) error {
	return uc.repo.BuildCommentIndexCache(ctx, param)
}

// LikeComment 评论点赞
func (uc *CommentUsecase) LikeComment(ctx context.Context, id uint64, like int, memberId uint64) error {
	comment := &Comment{
		Id: id,
		Like: like,
		MemberId: memberId,
	}
	return uc.repo.UpdateLikeNum(ctx, comment)
}
