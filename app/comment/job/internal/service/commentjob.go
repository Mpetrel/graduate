package service

import (
	pb "base-service/api/comment/job/v1"
	"base-service/app/comment/job/internal/biz"
	"base-service/app/comment/job/internal/data"
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/log"
)

type CommentJobService struct {
	pb.UnimplementedCommentJobServer
	uc *biz.CommentUsecase
	log *log.Helper
	kafka data.Queue
}

type SaveCommentMessage struct {
	Subject biz.CommentSubject
	Comment biz.Comment
}


type CommentIndexCacheMessage struct {
	ObjId uint64
	ObjType int
	Page int
	Size int
}


func NewCommentJobService(uc *biz.CommentUsecase, logger log.Logger, d *data.Data) *CommentJobService {
	service := &CommentJobService{
		uc: uc,
		log: log.NewHelper(logger),
		kafka: d.Kafka,
	}
/*	err := service.kafka.Subscribe("comment", func(ctx context.Context, message *data.Message) {
		service.log.Infow("type", "subscribe", "topic", message.Topic, "value", string(message.Value))
	})
	if err != nil {
		service.log.Errorf("subscribe kafka error: %v\n", err)
		return nil
	}*/
	go service.kafkaQueue()
	go service.cacheBuildQueue()
	return &CommentJobService{}
}


func (s *CommentJobService) kafkaQueue() {
	consumer, err := s.kafka.SubscribeChan("comment_chan", 256)
	if err != nil {
		s.log.Errorf("subscribe kafka chan error: %v\n", err)
		return
	}
	for msg := range consumer.Receive() {
		s.log.Infow("type", "subscribe channel", "topic", msg.Topic, "value", string(msg.Value))
		// 处理消息
		var param SaveCommentMessage
		err = json.Unmarshal(msg.Value, &param)
		if err != nil {
			s.log.Errorf("unmarshal message err: %v\n", err)
			continue
		}
		err = s.uc.CreateComment(context.Background(), &param.Subject, &param.Comment)
		if err != nil {
			s.log.Errorf("save comment err: %v\n", err)
		}
	}
}

// 构建缓存消息
func (s *CommentJobService) cacheBuildQueue() {
	consumer, err := s.kafka.SubscribeChan("comment-index-list-cache", 256)
	if err != nil {
		s.log.Errorf("subscribe kafka chan comment-index-list-cache error: %v\n", err)
		return
	}
	for msg := range consumer.Receive() {
		s.log.Infow("type", "subscribe channel", "topic", msg.Topic, "value", string(msg.Value))
		// 处理消息
		var param CommentIndexCacheMessage
		err = json.Unmarshal(msg.Value, &param)
		if err != nil {
			s.log.Errorf("unmarshal message err: %v\n", err)
			continue
		}
		err = s.uc.BuildCommentIndexCache(context.Background(), biz.CommentIndexCache{
			ObjId:   param.ObjId,
			ObjType: param.ObjType,
			Page:    param.Page,
			Size:    param.Size,
		})
		if err != nil {
			s.log.Errorf("build comment cache err: %v\n", err)
		}
	}
}


