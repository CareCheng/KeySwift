package model

import "time"

// SchemaRevision 记录当前数据库基线的结构变更指纹。
//
// 开发期不做旧库兼容迁移，该模型仅用于确认数据库已按当前 schema 构建。
type SchemaRevision struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	SchemaKey    string     `gorm:"type:varchar(50);uniqueIndex:idx_schema_revision_unique" json:"schema_key"`
	Version      string     `gorm:"type:varchar(50);uniqueIndex:idx_schema_revision_unique" json:"version"`
	Name         string     `gorm:"type:varchar(200)" json:"name"`
	Direction    string     `gorm:"type:varchar(30);uniqueIndex:idx_schema_revision_unique" json:"direction"`
	Checksum     string     `gorm:"type:varchar(128)" json:"checksum"`
	Status       string     `gorm:"type:varchar(30)" json:"status"`
	AppVersion   string     `gorm:"type:varchar(100)" json:"app_version"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (SchemaRevision) TableName() string {
	return "schema_revisions"
}
