package biz

import (
	"base-service/app/baseapp/interface/internal/pkg/token"
	"context"
	mapset "github.com/deckarep/golang-set"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
)

type CommentSubject struct {
	Id uint64
	ObjId uint64
	ObjType int
	MemberId uint64
	Count int
	RootCount int
	AllCount int
	State int8
	CreatedAt int64
	UpdatedAt int64
}

type Comment struct {
	Id uint64
	MemberId uint64
	Nickname string
	Avatar string
	Root uint64
	Parent uint64
	ParentMemberId uint64
	ParentNickname string
	ParentAvatar string
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
	CreatedAt int64
	UpdatedAt int64
	Replies []*Comment
}

type CommentRepo interface {
	GetCommentSubject(ctx context.Context, subject *CommentSubject) error
	SaveComment(ctx context.Context, subject *CommentSubject, comment *Comment) error
	GetComment(ctx context.Context, id uint64) (*Comment, error)
	GetCommentList(ctx context.Context, subject *CommentSubject, page, size, replyCount int) ([]*Comment, error)
	GetReplyList(ctx context.Context, rootId uint64, page, size int) ([] *Comment, error)
	LikeComment(ctx context.Context, comment *Comment) error
	GetLikeItem(ctx context.Context, memberId uint64, commentIds []uint64) (map[uint64]bool, error)
}

type CommentUsecase struct {
	repo CommentRepo
	accountRepo AccountRepo
	log *log.Helper
}

func NewCommentUsecase(repo CommentRepo, accountRepo AccountRepo, logger log.Logger) *CommentUsecase {
	return &CommentUsecase{
		repo: repo,
		accountRepo: accountRepo,
		log: log.NewHelper(logger),
	}
}

func (uc *CommentUsecase) GetCommentSubject(ctx context.Context, subject *CommentSubject) error {
	return uc.repo.GetCommentSubject(ctx, subject)
}

func (uc *CommentUsecase) SaveComment(ctx context.Context, subject *CommentSubject, comment *Comment) error {
	return uc.repo.SaveComment(ctx, subject, comment)
}

func (uc *CommentUsecase) GetComment(ctx context.Context, id uint64) (*Comment, error) {
	comment, err := uc.repo.GetComment(ctx, id)
	if err != nil {
		return nil, err
	}
	accounts, err := uc.accountRepo.ListByIds(ctx, []uint64{comment.MemberId})
	if err != nil {
		return nil, err
	}
	if len(accounts) > 0 {
		comment.Nickname = accounts[0].Nickname
		comment.Avatar = accounts[0].Avatar
	}
	return comment, nil
}

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

func (uc *CommentUsecase) GetReplyList(ctx context.Context, rootId uint64, page, size int) ([]*Comment, error) {
	comments, err := uc.repo.GetReplyList(ctx, rootId, page, size)
	if err != nil {
		return nil, err
	}
	accounts, err := uc.accountRepo.ListByIds(ctx, getCommentMemberIds(comments))
	if err != nil {
		return nil, err
	}
	// 如果有用户信息，查询用户是否点过赞
	likedMap := make(map[uint64]bool)
	if uid, err := token.ExtractUid(ctx); err == nil {
		temp, err := uc.repo.GetLikeItem(ctx, uid, getCommentIds(comments))
		if err != nil {
			uc.log.Errorf("查询用户是否点赞失败！%v \n", err)
		} else {
			likedMap = temp
		}
	}
	concatMemberInfo(comments, accounts, likedMap)
	return comments, nil
}

func (uc *CommentUsecase) LikeComment(ctx context.Context, comment *Comment) error {
	if comment.Like >= 0 {
		comment.Like = 1
	} else {
		comment.Like = -1
	}
	return uc.repo.LikeComment(ctx, comment)
}


func getCommentIds(comments []*Comment) []uint64 {
	result := make([]uint64, len(comments))
	for i := range comments {
		result[i] = comments[i].Id
	}
	return result
}

func getCommentMemberIds(comments []*Comment) []uint64 {
	idSet := mapset.NewSet()
	for i := range comments {
		idSet.Add(comments[i].MemberId)
		idSet.Add(comments[i].ParentMemberId)
		for j := range comments[i].Replies {
			idSet.Add(comments[i].Replies[j].MemberId)
			idSet.Add(comments[i].Replies[j].ParentMemberId)
		}
	}
	result := idSet.ToSlice()
	ids := make([]uint64, len(result))
	for i := range result {
		ids[i] = result[i].(uint64)
	}
	return ids
}

func concatMemberInfo(comments []*Comment, accounts []*Account, likedMap map[uint64]bool) {
	accountMap := make(map[uint64]*Account)
	for i := range accounts {
		accountMap[accounts[i].Id] = accounts[i]
	}
	for i := range comments {
		comments[i].Liked = likedMap[comments[i].Id]
		if account := accountMap[comments[i].MemberId]; account != nil {
			comments[i].Nickname = account.Nickname
			comments[i].Avatar = account.Avatar
		}
		if account := accountMap[comments[i].ParentMemberId]; account != nil {
			comments[i].ParentNickname = account.Nickname
			comments[i].ParentAvatar = account.Avatar
		}
		// sub comments
		for j := range comments[i].Replies {
			if account := accountMap[comments[i].Replies[j].MemberId]; account != nil {
				comments[i].Replies[j].Nickname = account.Nickname
				comments[i].Replies[j].Avatar = account.Avatar
			}
			if account := accountMap[comments[i].Replies[j].ParentMemberId]; account != nil {
				comments[i].Replies[j].ParentNickname = account.Nickname
				comments[i].Replies[j].ParentAvatar = account.Avatar
			}
		}
	}
}

