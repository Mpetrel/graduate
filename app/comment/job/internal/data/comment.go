package data

import (
	"base-service/app/comment/job/internal/biz"
	"base-service/pkg/orm"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"time"
)

type CommentSubject struct {
	orm.Model
	ObjId uint64
	ObjType int
	MemberId uint64
	Count int `gorm:"default:0"`
	RootCount int `gorm:"default:0"`
	AllCount int `gorm:"default:0"`
	State int8 `gorm:"default:0"`
}

type CommentIndex struct {
	orm.Model
	SubjectId uint64
	ObjId uint64
	ObjType int
	MemberId uint64
	Root uint64 `gorm:"default:0"`
	Parent uint64 `gorm:"default:0"`
	Floor int `gorm:"default:0"`
	Count int `gorm:"default:0"`
	RootCount int `gorm:"default:0"`
	Like int `gorm:"default:0"`
	Hate int `gorm:"default:0"`
	State int8 `gorm:"default:0"`
	Content CommentContent `gorm:"foreignKey:Id"`
}

func (i CommentIndex) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

type CommentContent struct {
	orm.Model
	AtMemberIds string
	Ip string
	Platform int8
	Device string
	Message string
	Meta string
}

type CommentLike struct {
	MemberId uint64 `gorm:"primaryKey;autoIncrement:false"`
	CommentId uint64 `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// 表明定义

func (CommentSubject) TableName() string {
	return "comment_subject"
}

func (CommentIndex) TableName() string {
	return "comment_index"
}

func (CommentContent) TableName() string {
	return "comment_content"
}

func (CommentLike) TableName() string {
	return "comment_like"
}

// repo 定义 实现

type commentRepo struct {
	data *Data
	log *log.Helper
}

func NewCommentRepo(data *Data, logger log.Logger) biz.CommentRepo {
	return &commentRepo{
		data: data,
		log: log.NewHelper(logger),
	}
}

func (c commentRepo) CreateSubject(ctx context.Context, subject *biz.CommentSubject) error {
	cs := CommentSubject{
		ObjId:     subject.ObjId,
		ObjType:   subject.ObjType,
		MemberId:  subject.MemberId,
		Count:     subject.Count,
		RootCount: subject.RootCount,
		AllCount:  subject.AllCount,
		State:     subject.State,
	}
	cs.Id = orm.NextId()
	result := c.data.db.WithContext(ctx).Create(&cs)
	subject.Id = cs.Id
	return result.Error
}


func (c commentRepo) queryOrCreateSubject(ctx context.Context, objId uint64, objType int, memberId uint64) (*CommentSubject, error) {
	sbj := CommentSubject{
		ObjId: objId,
		ObjType: objType,
	}
	result := c.data.db.WithContext(ctx).
		Where(CommentSubject{ObjId: objId, ObjType: objType}).
		First(&sbj)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 插入新数据
			sql := `INSERT INTO comment_subject(id, obj_id, obj_type, member_id, count, root_count, all_count, state, 
					created_at, updated_at) SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ? FROM DUAL 
			WHERE NOT EXISTS ( SELECT * FROM comment_subject WHERE obj_id=? AND obj_type=? );`
			t := time.Now()
			id := orm.NextId()
			saveResult := c.data.db.WithContext(ctx).Exec(sql, id, objId, objType, memberId, 0, 0, 0, 0,
				t, t, objId, objType)
			if saveResult.Error != nil {
				return nil, saveResult.Error
			}
			if saveResult.RowsAffected == 1 {
				sbj.Id = id
				sbj.CreatedAt = t
				return &sbj, nil
			} else {
				// 不为1则已经有数据插入，再次查询并返回
				result = c.data.db.WithContext(ctx).
					Where(CommentSubject{ObjId: objId, ObjType: objType}).
					First(&sbj)
			}
		} else {
			return nil, result.Error
		}
	}
	return &sbj, nil
}


func (c commentRepo) BuildCommentIndexCache(ctx context.Context, param biz.CommentIndexCache) error {
	offset := getOffset(param.Page, param.Size)
	var indexList []*CommentIndex
	indexResult := c.data.db.WithContext(ctx).
		Joins("Content").
		Where("obj_id = ? AND obj_type = ? AND root = ?", param.ObjId, param.ObjType, 0).
		Order("floor desc").
		Limit(param.Size).
		Offset(offset).
		Find(&indexList)
	if indexResult.Error != nil {
		return indexResult.Error
	}
	zList := make([]*redis.Z, len(indexList))
	for i := range indexList {
		zList[i] = &redis.Z{
			Score: float64(indexList[i].Floor),
			Member: indexList[i],
		}
	}
	result := c.data.redisDB.ZAdd(ctx, fmt.Sprintf("ci:%d:%d", param.ObjId, param.ObjType), zList...)
	return result.Err()
}


func (c commentRepo) SaveComment(ctx context.Context, subject *biz.CommentSubject, comment *biz.Comment) error {
	// 查询subject
	savedSbj, err := c.queryOrCreateSubject(ctx, subject.ObjId, subject.ObjType, subject.MemberId)
	if err != nil {
		return err
	}
	// 插入内容
	content := CommentContent{
		AtMemberIds: comment.AtMemberIds,
		Ip:          comment.Ip,
		Platform:    comment.Platform,
		Device:      comment.Device,
		Message:     comment.Message,
		Meta:        comment.Meta,
	}
	content.Id = comment.Id
	r := c.data.db.WithContext(ctx).
		Create(&content)
	if r.Error != nil {
		return r.Error
	}
	// 事务更新subject和index表
	err = c.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新subject
		updateFields := map[string]interface{} {
			"count": gorm.Expr("count + ?", 1),
			"all_count": gorm.Expr("all_count + ?", 1),
		}
		if comment.Root == 0 {
			updateFields["root_count"] = gorm.Expr("root_count + ?", 1)
		}
		result := tx.Debug().Model(&savedSbj).
			Updates(updateFields)
		if result.Error != nil {
			return result.Error
		}
		// update finish and get new data
		result = tx.First(&savedSbj, savedSbj.Id)
		if result.Error != nil {
			return result.Error
		}
		// 非根评论，楼层为根评论的楼层计数, 同时需要更新该评论的根评论楼层数
		floor := savedSbj.RootCount
		if comment.Root != 0 {
			rootIndex := CommentIndex{}
			rootIndex.Id = comment.Root
			result = tx.Model(&rootIndex).
				Updates(map[string]interface{}{
					"count": gorm.Expr("count + ?", 1),
					"root_count": gorm.Expr("root_count + ?", 1),
				})
			if result.Error != nil {
				return result.Error
			}
			// 查询最新楼层
			result = tx.First(&rootIndex, comment.Root)
			if result.Error != nil {
				return result.Error
			}
			floor = rootIndex.RootCount
		}
		// 插入index表
		ci := CommentIndex{
			SubjectId: savedSbj.Id,
			ObjId:     subject.ObjId,
			ObjType:   subject.ObjType,
			MemberId:  comment.MemberId,
			Root:      comment.Root,
			Parent:    comment.Parent,
			Floor:     floor,
		}
		ci.Id = content.Id
		comment.Id = content.Id
		return tx.Create(&ci).Error
	})
	if err == nil {
		go c.UpdateCommentIndexCache(*comment)
	}
	return err
}

func (c commentRepo) UpdateCommentIndexCache(comment biz.Comment) {
	// 查询刚保存的信息
	var ci CommentIndex
	result := c.data.db.First(&ci, comment.Id)
	if result.Error != nil {
		return
	}
	c.data.redisDB.ZAdd(context.Background(), fmt.Sprintf("ci:%d:%d", ci.ObjId, ci.ObjType), &redis.Z{
		Score: float64(ci.Floor),
		Member: ci,
	})
}

func (c commentRepo) UpdateLikeNum(ctx context.Context, comment *biz.Comment) error {
	if comment.Like == 0 {
		return nil
	} else if comment.Like > 1 {
		comment.Like = 1
	} else if comment.Like < -1 {
		comment.Like = -1
	}
	return c.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询是否有记录
		var likedRecord CommentLike
		result := tx.
			Where("member_id = ? AND comment_id = ?", comment.MemberId, comment.Id).
			First(&likedRecord)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			return result.Error
		} else if result.Error == gorm.ErrRecordNotFound && comment.Like == -1 {
			return nil // 不存在记录，则不能减一
		} else if result.Error == nil && comment.Like == 1 {
			return nil // 存在记录时，不能加1
		}
		// 更新like数量
		var ci CommentIndex
		ci.Id = comment.Id
		result = tx.Model(&ci).Updates(orm.UpdateFields{
			"like": gorm.Expr("`like` + ?", comment.Like),
		})
		if result.Error != nil {
			return result.Error
		}
		// 插入或删除记录
		if comment.Like == 1 {
			like := CommentLike{
				CommentId: comment.Id,
				MemberId: comment.MemberId,
			}
			result = tx.Create(&like)
		} else {
			result = tx.Where("member_id = ? AND comment_id = ?", comment.MemberId, comment.Id).
				Unscoped().Delete(CommentLike{})
		}
		return result.Error
	})
}

