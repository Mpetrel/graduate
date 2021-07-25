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

type CommentRepo interface {
	CreateSubject(ctx context.Context, subject *CommentSubject) error
	ListCommentSubject(ctx context.Context, objIds []uint64, objType int) ([]*CommentSubject, error)
	GetSubjectByObj(ctx context.Context, subject *CommentSubject) error
	GetCommentList(ctx context.Context, subject *CommentSubject, page, size, replyCount int) ([]*Comment, error)
	GetReplyList(ctx context.Context, rootId uint64, page, size int) ([] *Comment, error)
	SaveComment(ctx context.Context, subject *CommentSubject, comment *Comment) error
	GetCommentIndex(ctx context.Context, id uint64) (comment *Comment, err error)
	UpdateLikeNum(ctx context.Context, comment *Comment) error
	UpdateHateNum(ctx context.Context, comment *Comment) error
	GetLikedComment(ctx context.Context, memberId uint64, commentIds []uint64) ([]*LikeItem, error)
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

func (uc *CommentUsecase) GetSubject(ctx context.Context, subject *CommentSubject) error {
	return uc.repo.GetSubjectByObj(ctx, subject)
}

func (uc *CommentUsecase) ListCommentSubject(ctx context.Context, objIds []uint64, objType int) ([]*CommentSubject, error) {
	return uc.repo.ListCommentSubject(ctx, objIds, objType)
}

// GetComments 查询某个主题下的评论
// @param replyCount int 子评论条数
func (uc *CommentUsecase) GetComments(ctx context.Context, subject *CommentSubject, page int, size int, replyCount int) (comments []*Comment, err error) {
	// 查询相关评论
	comments, err = uc.repo.GetCommentList(ctx, subject, page, size, replyCount)
	return
}

func (uc *CommentUsecase) GetCommentById(ctx context.Context, id uint64) (*Comment, error) {
	return uc.repo.GetCommentIndex(ctx, id)
}


// GetReplies 查询一条评论下的回复
func (uc *CommentUsecase) GetReplies(ctx context.Context, rootId uint64, page int, size int) (comments []*Comment, err error) {
	comments, err = uc.repo.GetReplyList(ctx, rootId, page, size)
	return
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


func (uc *CommentUsecase) GetLikedComment(ctx context.Context, memberId uint64, commentIds []uint64) ([]*LikeItem, error) {
	return uc.repo.GetLikedComment(ctx, memberId, commentIds)
}