package MegaGormAudit

import (
	"gorm.io/plugin/soft_delete"
	"time"
)

type AuditableModel struct {
	ID              uint            `gorm:"primarykey" auditable:"true"`
	AuditParentID   *uint           `gorm:"default:null"`
	AuditParent     *AuditableModel `gorm:"foreignKey:AuditParentID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       soft_delete.DeletedAt
	LastChangedUser string
}
