package MegaGormAudit

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"testing"
	"time"
)

func TestAuditableModel_BeforeDelete(t *testing.T) {
	type fields struct {
		ID              uint
		AuditParentID   *uint
		AuditParent     *AuditableModel
		CreatedAt       time.Time
		UpdatedAt       time.Time
		DeletedAt       soft_delete.DeletedAt
		LastChangedUser string
	}
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &AuditableModel{
				ID:              tt.fields.ID,
				AuditParentID:   tt.fields.AuditParentID,
				AuditParent:     tt.fields.AuditParent,
				CreatedAt:       tt.fields.CreatedAt,
				UpdatedAt:       tt.fields.UpdatedAt,
				DeletedAt:       tt.fields.DeletedAt,
				LastChangedUser: tt.fields.LastChangedUser,
			}
			if err := u.BeforeDelete(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("BeforeDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMegaGormAuditPlugin_Initialize(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MegaGormAuditPlugin{}
			if err := a.Initialize(tt.args.db); (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMegaGormAuditPlugin_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MegaGormAuditPlugin{}
			if got := a.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createDatabase(config *gorm.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(""), config)
	if err != nil {
		return nil, err
	}
	return db, nil
}
