package service

import (
	"base-service/app/baseapp/interface/internal/biz"
	"base-service/app/baseapp/interface/internal/pkg/token"
	"context"
	"github.com/go-kratos/kratos/v2/log"

	pb "base-service/api/baseapp/interface/v1"
)

type BaseappInterfaceService struct {
	pb.UnimplementedBaseappInterfaceServer
	uc *biz.CommentUsecase
	accountUC *biz.AccountUsecase
	log *log.Helper
}

func NewBaseappInterfaceService(
		uc *biz.CommentUsecase,
		accountUC *biz.AccountUsecase,
		logger log.Logger) *BaseappInterfaceService {
	return &BaseappInterfaceService{
		uc: uc,
		accountUC: accountUC,
		log: log.NewHelper(logger),
	}
}
func (s *BaseappInterfaceService) GetCommentSubject(ctx context.Context, req *pb.GetCommentSubjectRequest) (*pb.GetCommentSubjectReply, error) {
	if req.ObjId <= 0 || req.ObjType <= 0{
		return nil, pb.ErrorContentMissing("invalid params")
	}
	subject := &biz.CommentSubject{ObjId: req.ObjId, ObjType: int(req.ObjType)}
	err := s.uc.GetCommentSubject(ctx, subject)
	return &pb.GetCommentSubjectReply{
		Id:        subject.Id,
		ObjId:     subject.ObjId,
		ObjType:   int32(subject.ObjType),
		MemberId:  subject.MemberId,
		Count:     int32(subject.Count),
		RootCount: int32(subject.RootCount),
		AllCount:  int32(subject.AllCount),
		State: 	   int32(subject.State),
		CreatedAt: subject.CreatedAt,
	}, err
}


func (s *BaseappInterfaceService) SaveComment(ctx context.Context, req *pb.SaveCommentRequest) (*pb.SaveCommentReply, error) {
	uid, err := token.ExtractUid(ctx)
	if err != nil {
		return nil, pb.ErrorUNAUTHORIZED("unauthorized")
	}
	subject := &biz.CommentSubject{
		ObjType: int(req.ObjType),
		ObjId:   req.ObjId,
	}
	comment := &biz.Comment{
		MemberId:    uid,
		Root:        req.Root,
		Parent:      req.Parent,
		State:       0,
		AtMemberIds: "",
		Ip:          "",
		Platform: 	 0,
		Device:      "unknown",
		Message:     req.Content,
		Meta:        req.Meta,
	}
	err = s.uc.SaveComment(ctx, subject, comment)
	return &pb.SaveCommentReply{
		Id: comment.Id,
	}, err
}

func (s *BaseappInterfaceService) GetComment(ctx context.Context, req *pb.GetCommentRequest) (*pb.GetCommentReply, error) {
	result, err := s.uc.GetComment(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetCommentReply{Comment: createCommentData(result)}, nil
}

func (s *BaseappInterfaceService) GetCommentList(ctx context.Context, req *pb.GetCommentListRequest) (*pb.GetCommentListReply, error) {
	list, err := s.uc.GetCommentList(ctx, &biz.CommentSubject{
		ObjId:   req.ObjId,
		ObjType: int(req.ObjType),
	}, int(req.Page), int(req.Size), int(req.Reply))
	if err != nil {
		s.log.Errorf("rpc get comment list failed: %v", err)
		return nil, pb.ErrorInfoNotFound("comment not found, obj_id: %d", req.ObjId)
	}
	comments := make([]*pb.CommentData, len(list))
	for i := range list {
		comments[i] = createCommentData(list[i])
		for j := range list[i].Replies {
			comments[i].Replies = append(comments[i].Replies, createCommentData(list[i].Replies[j]))
		}
	}

	return &pb.GetCommentListReply{
		Comments: comments,
	}, nil
}

func (s *BaseappInterfaceService) GetReplyList(ctx context.Context, req *pb.GetReplyListRequest) (*pb.GetCommentListReply, error) {
	list, err := s.uc.GetReplyList(ctx, req.RootId, int(req.Page), int(req.Size))
	if err != nil {
		return nil, pb.ErrorInfoNotFound("comment reply for %d not found", req.RootId)
	}
	replies := make([]*pb.CommentData, len(list))
	for i := range list {
		replies[i] = createCommentData(list[i])
	}
	return &pb.GetCommentListReply{
		Comments: replies,
	}, nil
}

func (s *BaseappInterfaceService) LikeComment(ctx context.Context, req *pb.LikeCommentRequest) (*pb.LikeCommentReply, error) {
	uid, err := token.ExtractUid(ctx)
	if err != nil {
		return nil, pb.ErrorUNAUTHORIZED("unauthorized")
	}
	err = s.uc.LikeComment(ctx, &biz.Comment{
		Id:       req.Id,
		MemberId: uid,
		Like: 	  int(req.Like),
	})
	return &pb.LikeCommentReply{}, err
}


func (s *BaseappInterfaceService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	account, tokenStr, err := s.accountUC.Login(ctx, req.Account, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginReply{
		Token: tokenStr,
		Account: &pb.AccountInfo{
			Id:       account.Id,
			Nickname: account.Nickname,
			Avatar:   account.Avatar,
		},
	}, err
}


func createCommentData(comment *biz.Comment) *pb.CommentData {
	return &pb.CommentData{
		Id:          comment.Id,
		MemberId:    comment.MemberId,
		Nickname: 	 comment.Nickname,
		Avatar: 	 comment.Avatar,
		Root:        comment.Root,
		Parent:      comment.Parent,
		ParentMemberId: comment.ParentMemberId,
		ParentNickname: comment.ParentNickname,
		ParentAvatar: comment.ParentAvatar,
		Floor:       int32(comment.Floor),
		Count:       int32(comment.Count),
		RootCount:   int32(comment.RootCount),
		Like:        int32(comment.Like),
		Liked: 		 comment.Liked,
		Hate:        int32(comment.Hate),
		State:       int32(comment.State),
		AtMemberIds: comment.AtMemberIds,
		Ip:          comment.Ip,
		Platform: 	 int32(comment.Platform),
		Device:      comment.Device,
		Message:     comment.Message,
		Meta:        comment.Meta,
		CreateAt:    comment.CreatedAt,
		UpdatedAt:   comment.UpdatedAt,
		Replies:     make([]*pb.CommentData, 0),
	}
}