package service

import (
	"base-service/app/comment/service/internal/biz"
	"context"
	"github.com/go-kratos/kratos/v2/log"

	pb "base-service/api/comment/service/v1"
)

type CommentService struct {
	pb.UnimplementedCommentServer
	uc *biz.CommentUsecase
	log *log.Helper
}

func NewCommentService(uc *biz.CommentUsecase, logger log.Logger) *CommentService {
	return &CommentService{
		uc: uc,
		log: log.NewHelper(logger),
	}
}


func (s *CommentService) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CreateCommentReply, error) {
	subject := &biz.CommentSubject{
		ObjType: int(req.ObjType),
		ObjId:   req.ObjId,
	}
	comment := &biz.Comment{
		MemberId:    req.MemberId,
		Root:        req.Root,
		Parent:      req.Parent,
		State:       int8(req.State),
		AtMemberIds: req.AtMemberIds,
		Ip:          req.Ip,
		Platform: 	 int8(req.Platform),
		Device:      req.Device,
		Message:     req.Message,
		Meta:        req.Meta,
	}
	err := s.uc.CreateComment(ctx, subject, comment)
	return &pb.CreateCommentReply{Id: comment.Id}, err
}

func (s *CommentService) GetComment(ctx context.Context, req *pb.GetCommentRequest) (*pb.GetCommentReply, error) {
	result, err := s.uc.GetCommentById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetCommentReply{Comment: createCommentData(result)}, nil
}

func (s *CommentService) LikeComment(ctx context.Context, req *pb.LikeCommentRequest) (*pb.LikeCommentReply, error) {
	err := s.uc.LikeComment(ctx, req.Id, int(req.Like), req.MemberId)
	return &pb.LikeCommentReply{}, err
}

func (s *CommentService) DeleteComment(context.Context, *pb.DeleteCommentRequest) (*pb.DeleteCommentReply, error) {
	return &pb.DeleteCommentReply{}, nil
}

func (s *CommentService) ListComment(ctx context.Context, req *pb.ListCommentRequest) (*pb.ListCommentReply, error) {
	comments, err := s.uc.GetComments(ctx, &biz.CommentSubject{
		ObjType: int(req.ObjType),
		ObjId: req.ObjId,
	}, int(req.Page), int(req.Size), int(req.ReplyCount))
	if err != nil {
		s.log.Errorf("get comments failed: %v", err)
		return nil, err
	}
	result := make([]*pb.CommentData, len(comments))
	for i, v := range comments {
		result[i] = createCommentData(v)
		// 迭代创建子评论
		if len(v.Replies) > 0 {
			for _, sub := range v.Replies {
				result[i].Replies = append(result[i].Replies, createCommentData(sub))
			}
		}
	}
	return &pb.ListCommentReply{
		Comments: result,
	}, nil
}

func (s *CommentService) ListSubComment(ctx context.Context, req *pb.ListSubCommentRequest) (*pb.ListCommentReply, error) {
	comments, err := s.uc.GetReplies(ctx, req.RootId, int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}
	result := make([]*pb.CommentData, len(comments))
	for i := range comments {
		result[i] = createCommentData(comments[i])
	}
	return &pb.ListCommentReply{Comments: result}, nil
}

func (s *CommentService) GetCommentSubject(ctx context.Context, req *pb.GetCommentSubjectRequest) (*pb.GetCommentSubjectReply, error) {
	subject := &biz.CommentSubject{
		ObjType: int(req.ObjType),
		ObjId:   req.ObjId,
	}
	err := s.uc.GetSubject(ctx, subject)
	if err != nil {
		s.log.Errorf("get comment subject failed, subjectId: %d, subjectType: %s\nerr: %v", req.ObjId, req.ObjType, err)
		return nil, err
	}
	return &pb.GetCommentSubjectReply{
		Id:        subject.Id,
		ObjId:     subject.ObjId,
		ObjType:   int32(subject.ObjType),
		MemberId:  subject.MemberId,
		Count:     int32(subject.Count),
		RootCount: int32(subject.RootCount),
		AllCount:  int32(subject.AllCount),
		State: 	   int32(subject.State),
		CreatedAt: subject.CreatedAt.Unix(),
	}, nil
}

func (s *CommentService) ListCommentSubject(ctx context.Context, req *pb.ListCommentSubjectRequest) (*pb.ListCommentSubjectReply, error) {
	subjects, err := s.uc.ListCommentSubject(ctx, req.Ids, int(req.ObjType))
	if err != nil {
		return nil, err
	}
	result := make([]*pb.ListCommentSubjectReply_CommentSubject, len(subjects))
	for i, item := range subjects {
		cs := &pb.ListCommentSubjectReply_CommentSubject{
			Id:        item.Id,
			ObjId:     item.ObjId,
			ObjType:   int32(item.ObjType),
			MemberId:  item.MemberId,
			Count:     int32(item.Count),
			RootCount: int32(item.RootCount),
			AllCount:  int32(item.AllCount),
			State: 	   int32(item.State),
			CreatedAt: item.CreatedAt.Unix(),
		}
		result[i] = cs
	}
	return &pb.ListCommentSubjectReply{CommentSubjects: result}, nil
}

func (s *CommentService) GetCommentLiked(ctx context.Context, req *pb.GetCommentLikedRequest) (*pb.GetCommentLikedReply, error) {
	result, err := s.uc.GetLikedComment(ctx, req.MemberId, req.CommentId)
	if err != nil {
		return nil, err
	}
	items := make([]*pb.GetCommentLikedReply_LikedItem, len(result))
	for i := range result {
		items[i] = &pb.GetCommentLikedReply_LikedItem{
			CommentId: result[i].CommentId,
			Like:      result[i].Like,
		}
	}
	return &pb.GetCommentLikedReply{LikedItems: items}, nil
}

func createCommentData(comment *biz.Comment) *pb.CommentData {
	return &pb.CommentData{
		Id:          comment.Id,
		MemberId:    comment.MemberId,
		Root:        comment.Root,
		Parent:      comment.Parent,
		ParentMemberId: comment.ParentMemberId,
		Floor:       int32(comment.Floor),
		Count:       int32(comment.Count),
		RootCount:   int32(comment.RootCount),
		Like:        int32(comment.Like),
		Hate:        int32(comment.Hate),
		State:       int32(comment.State),
		AtMemberIds: comment.AtMemberIds,
		Ip:          comment.Ip,
		Platform: 	 int32(comment.Platform),
		Device:      comment.Device,
		Message:     comment.Message,
		Meta:        comment.Meta,
		CreateAt:    comment.CreatedAt.Unix(),
		UpdatedAt:   comment.UpdatedAt.Unix(),
		Replies:     make([]*pb.CommentData, 0),
	}
}