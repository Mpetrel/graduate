// Package orm define common fields
package orm

import (
	"github.com/sony/sonyflake"
	"gorm.io/gorm"
	"log"
	"time"
)

type Model struct {
	Id uint64 `gorm:"primarykey;autoIncrement:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type UpdateFields =  map[string]interface{}

var sf *sonyflake.Sonyflake

func init() {
	sf = sonyflake.NewSonyflake(sonyflake.Settings{})
}

func NextId() uint64 {
	nextId, err := sf.NextID()
	if err != nil {
		log.Printf("generate sony flake id failed: %v", err)
	}
	return nextId
}