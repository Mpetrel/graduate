package data

import (
	"base-service/app/comment/service/internal/biz"
	"base-service/pkg/orm"
	"context"
	"encoding/json"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/go-kratos/kratos/v2/log"
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

func (i *CommentIndex) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, i)
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


func (c commentRepo) ListCommentSubject(ctx context.Context, objIds []uint64, objType int) ([]*biz.CommentSubject, error) {
	var subjectList []*CommentSubject
	result := c.data.db.WithContext(ctx).
		Where("obj_id IN ? AND obj_type = ?", objIds, objType).
		Find(&subjectList)
	if result.Error != nil {
		return nil, result.Error
	}
	subjects := make([]*biz.CommentSubject, len(subjectList))
	for i := range subjectList {
		bcs := &biz.CommentSubject{}
		copyCommentSubject(bcs, subjectList[i])
		subjects[i] = bcs
	}
	return subjects, nil
}

// GetSubjectByObj 查询评论主题, 不存在则插入数据并返回
func (c commentRepo) GetSubjectByObj(ctx context.Context, subject *biz.CommentSubject) error {
	sbj, err := c.queryOrCreateSubject(ctx, subject.ObjId, subject.ObjType, subject.MemberId)
	if err != nil {
		return err
	}
	copyCommentSubject(subject, sbj)
	return nil
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

func (c commentRepo) GetSubjectByObjBack(ctx context.Context, subject *biz.CommentSubject) error {
	var sbj CommentSubject
	result := c.data.db.WithContext(ctx).
		Where(CommentSubject{ObjId: subject.ObjId, ObjType: subject.ObjType}).
		Attrs(CommentSubject{Model: orm.Model{Id: orm.NextId()}}).
		FirstOrCreate(&sbj)
	subject.Id = sbj.Id
	subject.MemberId = sbj.MemberId
	subject.Count = sbj.Count
	subject.AllCount = sbj.AllCount
	subject.RootCount = sbj.RootCount
	subject.State = sbj.State
	subject.CreatedAt = sbj.CreatedAt
	subject.UpdatedAt = sbj.UpdatedAt
	return result.Error
}

// GetCommentList 根据subject_id 和 subject_type 查询相关评论
func (c commentRepo) GetCommentList(ctx context.Context, subject *biz.CommentSubject, page, size, replyCount int) ([]*biz.Comment, error) {
	offset := getOffset(page, size)
	var indexList []*CommentIndex
	indexListCache := false
	// 查询缓存数据
	cacheResult := c.data.redisDB.ZRange(ctx, fmt.Sprintf("ci:%d:%d", subject.ObjId, subject.ObjType), int64(offset), int64(offset + size))
	if cacheResult.Err() != nil {
		c.log.Errorf("get comment list read index list cache err: %v\n", cacheResult.Err())
	} else {
		err := cacheResult.ScanSlice(&indexList)
		if err != nil || indexList == nil {
			c.log.Errorf("get comment list scan index list cache err: %v\n", err)
		} else {
			indexListCache = true // read cache success
		}
	}
	// 查询主题下的根评论
	if !indexListCache {
		// 提交填充缓存的消息
		cicm, _ := json.Marshal(CommentIndexCacheMessage{
			ObjId: subject.ObjId,
			ObjType: subject.ObjType,
			Page: page,
			Size: size,
		})
		_ = c.data.Kafka.Send("comment-index-list-cache", string(cicm))
		// 回源数据库查询
		indexResult := c.data.db.WithContext(ctx).
			Joins("Content").
			Where("obj_id = ? AND obj_type = ? AND root = ?", subject.ObjId, subject.ObjType, 0).
			Order("floor desc").
			Limit(size).
			Offset(offset).
			Find(&indexList)
		if indexResult.Error != nil {
			return nil, indexResult.Error
		}
	}
	// 取出id，用于批量查询关联的内容, 同时组装根评论结果
	indexIds := make([]uint64, len(indexList))
	comments := make([]*biz.Comment, len(indexList))

	for i := range indexList {
		indexIds[i] = indexList[i].Id
		comments[i] = createComment(indexList[i])
	}
	// 如果replyCount 不为0 则需要查询子评论
	if replyCount > 0 {
		var subIndexList []*CommentIndex
		result := c.data.db.WithContext(ctx).
			Joins("Content").
			Where("root IN ? AND floor <= ?", indexIds, replyCount).
			Order("floor asc").
			Find(&subIndexList)
		if result.Error != nil {
			c.log.Errorf("get comment sub comment failed: %v", result.Error)
			return comments, nil
		}
		// 查询回复对象的作者id
		subParentId := mapset.NewSet()
		for i := range subIndexList {
			if subIndexList[i].Parent != 0 {
				subParentId.Add(subIndexList[i].Parent)
			}
		}
		// 查询回复
		var subParentIndex []*CommentIndex
		result = c.data.db.WithContext(ctx).
			Where("id IN ?", subParentId.ToSlice()).
			Find(&subParentIndex)
		if result.Error != nil {
			return comments, nil
		}

		parentMap := make(map[uint64]*CommentIndex)
		for i := range subParentIndex {
			parentMap[subParentIndex[i].Id] = subParentIndex[i]
		}

		for i := range comments {
			comments[i].Replies = findChild(comments[i].Id, subIndexList, parentMap)
		}
	}
	return comments, nil
}

// GetReplyList 查询一条评论下的回复列表
func (c commentRepo) GetReplyList(ctx context.Context, rootId uint64, page, size int) ([]*biz.Comment, error) {
	offset := getOffset(page, size)
	var indexList []*CommentIndex
	result := c.data.db.WithContext(ctx).
		Joins("Content").
		Where("root = ?", rootId).
		Order("floor desc").
		Limit(size).
		Offset(offset).
		Find(&indexList)
	if result.Error != nil {
		return nil, result.Error
	}

	// 查询parent
	parentIds := make([]uint64, len(indexList))
	for i := range indexList {
		parentIds[i] = indexList[i].Parent
	}
	var parentIndexList []*CommentIndex
	result = c.data.db.WithContext(ctx).
		Where("id IN ?", parentIds).
		Find(&parentIndexList)
	if result.Error != nil {
		return nil, result.Error
	}
	parentMap := make(map[uint64]*CommentIndex)
	for i := range parentIndexList {
		parentMap[parentIndexList[i].Id] = parentIndexList[i]
	}
	ret := make([]*biz.Comment, len(indexList))
	for i := range indexList {
		p := parentMap[indexList[i].Parent]
		if p != nil {
			ret[i] = createCommentWithParentMember(indexList[i], p.MemberId)
		} else {
			ret[i] = createComment(indexList[i])
		}
	}
	return ret, nil
}

func (c commentRepo) SaveComment(_ context.Context, subject *biz.CommentSubject, comment *biz.Comment) error {
	comment.Id = orm.NextId()
	paramByte, _ := json.Marshal(SaveCommentMessage{
		Subject: *subject,
		Comment: *comment,
	})
	err := c.data.Kafka.Send("comment_chan", string(paramByte))
	return err
}

func (c commentRepo) SaveCommentBack(ctx context.Context, subject *biz.CommentSubject, comment *biz.Comment) error {
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
	content.Id = orm.NextId()
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
	return err
}

func (c commentRepo) GetCommentIndex(ctx context.Context, id uint64) (comment *biz.Comment, err error) {
	var ci CommentIndex
	result := c.data.db.WithContext(ctx).Joins("Content").
		First(&ci, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return createComment(&ci), nil
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

func (c commentRepo) UpdateHateNum(ctx context.Context, comment *biz.Comment) error {
	result := c.data.db.WithContext(ctx).Updates(orm.UpdateFields{
		"hate": gorm.Expr("hate + ?", comment.Hate),
	})
	return result.Error
}

func (c commentRepo) GetLikedComment(ctx context.Context, memberId uint64, commentIds []uint64) ([]*biz.LikeItem, error) {
	var likeList []*CommentLike
	result := c.data.db.WithContext(ctx).
		Where("member_id = ? AND comment_id IN ?", memberId, commentIds).
		Find(&likeList)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return make([]*biz.LikeItem, 0), nil
		}
		return nil, result.Error
	}

	likeItems := make([]*biz.LikeItem, len(likeList))
	for i := range likeList {
		likeItems[i] = &biz.LikeItem{
			CommentId: likeList[i].CommentId,
			Like:      true,
		}
	}
	return likeItems, nil
}

func createComment(ci *CommentIndex) *biz.Comment {
	return &biz.Comment{
		Id:          ci.Id,
		MemberId:    ci.MemberId,
		Root:        ci.Root,
		Parent:      ci.Parent,
		Floor:       ci.Floor,
		Count:       ci.Count,
		RootCount:   ci.RootCount,
		Like:        ci.Like,
		Hate:        ci.Hate,
		State:       ci.State,
		AtMemberIds: ci.Content.AtMemberIds,
		Ip:          ci.Content.Ip,
		Platform:    ci.Content.Platform,
		Device:      ci.Content.Device,
		Message:     ci.Content.Message,
		Meta:        ci.Content.Meta,
		CreatedAt:   ci.Content.CreatedAt,
		UpdatedAt:   ci.Content.UpdatedAt,
		Replies:     make([]*biz.Comment, 0),
	}
}

func createCommentWithParentMember(ci *CommentIndex, parentMemberId uint64) *biz.Comment {
	c := createComment(ci)
	c.ParentMemberId = parentMemberId
	return c
}

func findChild(id uint64, target []*CommentIndex, parentMap map[uint64]*CommentIndex) []*biz.Comment {
	ret := make([]*biz.Comment, 0)
	for _, v := range target {
		if v.Root == id {
			parent := parentMap[v.Parent]
			var parentMemberId uint64 = 0
			if parent != nil {
				parentMemberId = parent.MemberId
			}
			ret = append(ret, createCommentWithParentMember(v, parentMemberId))
		}
	}
	return ret
}

func copyCommentSubject(subject *biz.CommentSubject, sbj *CommentSubject) {
	subject.Id = sbj.Id
	subject.ObjType = sbj.ObjType
	subject.ObjId = sbj.ObjId
	subject.MemberId = sbj.MemberId
	subject.Count = sbj.Count
	subject.AllCount = sbj.AllCount
	subject.RootCount = sbj.RootCount
	subject.State = sbj.State
	subject.CreatedAt = sbj.CreatedAt
	subject.UpdatedAt = sbj.UpdatedAt
}