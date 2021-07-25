package data

import (
	v1 "base-service/api/comment/service/v1"
	"base-service/app/baseapp/interface/internal/biz"
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type commentRepo struct {
	log *log.Helper
	data *Data
}

func NewCommentRepo(data *Data, logger log.Logger) biz.CommentRepo {
	return &commentRepo{
		log:  log.NewHelper(logger),
		data: data,
	}
}


func (c commentRepo) GetCommentSubject(ctx context.Context, subject *biz.CommentSubject) error {
	result, err := c.data.cc.GetCommentSubject(ctx, &v1.GetCommentSubjectRequest{
		ObjId:   subject.ObjId,
		ObjType: int32(subject.ObjType),
	})
	if err != nil {
		return err
	}
	subject.Id = result.Id
	subject.MemberId = result.MemberId
	subject.Count = int(result.Count)
	subject.AllCount = int(result.AllCount)
	subject.RootCount = int(result.RootCount)
	subject.State = int8(result.State)
	subject.CreatedAt = result.CreatedAt
	return nil
}


func (c commentRepo) SaveComment(ctx context.Context, subject *biz.CommentSubject, comment *biz.Comment) error {
	r, err := c.data.cc.CreateComment(ctx, &v1.CreateCommentRequest{
		ObjId:       subject.ObjId,
		ObjType:     int32(subject.ObjType),
		MemberId:    comment.MemberId,
		Root:        comment.Root,
		Parent:      comment.Parent,
		AtMemberIds: comment.AtMemberIds,
		Ip:          comment.Ip,
		Platform:    int32(comment.Platform),
		Device:      comment.Device,
		Message:     comment.Message,
		Meta:        comment.Meta,
	})
	if err != nil {
		return err
	}
	comment.Id = r.Id
	return nil
}


func (c commentRepo) GetComment(ctx context.Context, id uint64) (*biz.Comment, error) {
	result, err := c.data.cc.GetComment(ctx, &v1.GetCommentRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return toBizComment(result.Comment), err
}


func (c commentRepo) GetCommentList(ctx context.Context, subject *biz.CommentSubject, page, size, replyCount int) ([]*biz.Comment, error) {
	result, err := c.data.cc.ListComment(ctx, &v1.ListCommentRequest{
		ObjId:      subject.ObjId,
		ObjType:    int32(subject.ObjType),
		Page:       int32(page),
		Size:       int32(size),
		ReplyCount: int32(replyCount),
	})
	if err != nil {
		return nil, err
	}
	comments := make([]*biz.Comment, len(result.Comments))
	for i := range result.Comments {
		comments[i] = toBizComment(result.Comments[i])
		for j := range result.Comments[i].Replies {
			comments[i].Replies = append(comments[i].Replies, toBizComment(result.Comments[i].Replies[j]))
		}
	}
	return comments, nil
}

func (c commentRepo) GetReplyList(ctx context.Context, rootId uint64, page, size int) ([]*biz.Comment, error) {
	result, err := c.data.cc.ListSubComment(ctx, &v1.ListSubCommentRequest{
		RootId: rootId,
		Page:   int32(page),
		Size: 	int32(size),
	})
	if err != nil {
		return nil, err
	}
	replies := make([]*biz.Comment, len(result.Comments))
	for i := range result.Comments {
		replies[i] = toBizComment(result.Comments[i])
	}
	return replies, nil
}

func (c commentRepo) LikeComment(ctx context.Context, comment *biz.Comment) error {
	_, err := c.data.cc.LikeComment(ctx, &v1.LikeCommentRequest{
		Id:       comment.Id,
		Like: 	  int32(comment.Like),
		MemberId: comment.MemberId,
	})
	return err
}


func (c commentRepo) GetLikeItem(ctx context.Context, memberId uint64, commentIds []uint64) (map[uint64]bool, error) {
	result, err := c.data.cc.GetCommentLiked(ctx, &v1.GetCommentLikedRequest{
		MemberId:  memberId,
		CommentId: commentIds,
	})
	if err != nil {
		return nil, err
	}
	likedMap := make(map[uint64]bool)
	for _, val := range result.LikedItems {
		likedMap[val.CommentId] = val.Like
	}
	return likedMap, nil
}



func toBizComment(ci *v1.CommentData) *biz.Comment {
	return &biz.Comment{
		Id:          ci.Id,
		MemberId:    ci.MemberId,
		Root:        ci.Root,
		Parent:      ci.Parent,
		ParentMemberId: ci.ParentMemberId,
		Floor:       int(ci.Floor),
		Count:       int(ci.Count),
		RootCount:   int(ci.RootCount),
		Like:        int(ci.Like),
		Hate:        int(ci.Hate),
		State:       int8(ci.State),
		AtMemberIds: ci.AtMemberIds,
		Ip:          ci.Ip,
		Platform: 	 int8(ci.Platform),
		Device:      ci.Device,
		Message:     ci.Message,
		Meta:        ci.Meta,
		CreatedAt:   ci.CreateAt,
		UpdatedAt:   ci.UpdatedAt,
		Replies:     make([]*biz.Comment, 0),
	}
}
