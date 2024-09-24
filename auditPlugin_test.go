package MegaGormAudit

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"reflect"
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

func TestAuditPlugin_ModelWithoutAuditory(t *testing.T) {

	type normalModel struct {
		ID       uint64 `gorm:"primaryKey"`
		Name     string
		NickName string
		Address  string
	}

	tests := []struct {
		name        string
		model       *normalModel
		want        *normalModel
		afterCreate func(db *gorm.DB, model *normalModel) error
		wantErr     bool
	}{
		{
			name: "Success",
			model: &normalModel{
				ID:       0,
				Name:     "teste",
				NickName: "teste",
				Address:  "teste",
			},
			want: &normalModel{
				ID:       1,
				Name:     "teste",
				NickName: "teste",
				Address:  "teste",
			},
			afterCreate: nil,
			wantErr:     false,
		},
		{
			name: "Success, update",
			model: &normalModel{
				ID:       0,
				Name:     "teste",
				NickName: "teste",
				Address:  "teste",
			},
			want: &normalModel{
				ID:       1,
				Name:     "teste atualizado",
				NickName: "teste",
				Address:  "teste",
			},
			afterCreate: func(db *gorm.DB, model *normalModel) error {
				model.Name = "teste atualizado"
				return db.Updates(model).Error
			},
			wantErr: false,
		},
		{
			name: "Success, delete",
			model: &normalModel{
				ID:       0,
				Name:     "teste",
				NickName: "teste",
				Address:  "teste",
			},
			want: &normalModel{
				ID:       0,
				Name:     "",
				NickName: "",
				Address:  "",
			},
			afterCreate: func(db *gorm.DB, model *normalModel) error {
				return db.Delete(model).Error
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := createDatabase()
			if err != nil {
				t.Errorf("createDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = db.AutoMigrate(normalModel{})
			if err != nil {
				t.Errorf("AutoMigrate() error = %v", err)
				return
			}

			err = db.Create(tt.model).Error
			if err != nil {
				t.Errorf("Create() error = %v", err)
				return
			}

			if tt.afterCreate != nil {
				err = tt.afterCreate(db, tt.model)
				if err != nil {
					t.Errorf("afterCreate() error = %v", err)
					return
				}

			}

			row := &normalModel{}
			db.First(row)

			if !reflect.DeepEqual(row, tt.want) {
				t.Errorf("normalModel() = %v, want %v", row, tt.want)
			}
		})
	}
}

func TestAuditPlugin_ModelGormWithoutAuditory(t *testing.T) {

	type Player struct {
		gorm.Model
		Name     string
		NickName string
	}

	compareFunc := func(p1, p2 []Player) bool {
		if len(p1) != len(p2) {
			return false
		}

		for i := range p1 {
			if p1[i].Name != p2[i].Name || p1[i].NickName != p2[i].NickName || p1[i].ID != p2[i].ID || !reflect.DeepEqual(p1[i].DeletedAt, p2[i].DeletedAt) {
				return false
			}
		}
		return true
	}

	tests := []struct {
		name        string
		model       *Player
		want        []Player
		afterCreate func(db *gorm.DB, model *Player) error
		success     func(db *gorm.DB) bool
		wantErr     bool
	}{
		{
			name: "Success, Create",
			model: &Player{
				Model:    gorm.Model{},
				Name:     "teste",
				NickName: "teste",
			},
			want: []Player{
				{
					Model: gorm.Model{
						ID:        1,
						DeletedAt: deletedAtNull(),
					},
					Name:     "teste",
					NickName: "teste",
				},
			},
			success: func(db *gorm.DB) bool {
				var rows []Player
				db.Find(&rows)

				return len(rows) == 0
			},
			afterCreate: nil,
			wantErr:     false,
		},
		{
			name: "Success, update",
			model: &Player{
				Model:    gorm.Model{},
				Name:     "teste",
				NickName: "teste",
			},
			want: []Player{
				{
					Model: gorm.Model{
						ID:        1,
						DeletedAt: deletedAtNull(),
					},
					Name:     "teste atualizado",
					NickName: "teste",
				},
			},
			afterCreate: func(db *gorm.DB, model *Player) error {
				model.Name = "teste atualizado"
				return db.Updates(model).Error
			},
			success: func(db *gorm.DB) bool {
				var rows []Player
				db.Find(&rows)

				return len(rows) == 0
			},
			wantErr: false,
		},
		{
			name: "Success, delete",
			model: &Player{
				Model:    gorm.Model{},
				Name:     "teste",
				NickName: "teste",
			},
			want: []Player{},
			afterCreate: func(db *gorm.DB, model *Player) error {
				return db.Delete(model).Error
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			db, err := createDatabase()
			if err != nil {
				t.Errorf("createDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = db.AutoMigrate(Player{})
			if err != nil {
				t.Errorf("AutoMigrate() error = %v", err)
				return
			}

			err = db.Create(tt.model).Error
			if err != nil {
				t.Errorf("Create() error = %v", err)
				return
			}

			if tt.afterCreate != nil {
				err = tt.afterCreate(db, tt.model)
				if err != nil {
					t.Errorf("afterCreate() error = %v", err)
					return
				}

			}

			if tt.success != nil {
				if !tt.success(db) {
					t.Errorf("[]Player() = false, want true")
				}

			}

			var rows []Player
			db.Find(&rows)

			if !compareFunc(rows, tt.want) {
				t.Errorf("[]Player() = %v, want %v", rows, tt.want)
			}
			//
			//if !reflect.DeepEqual(rows, tt.want) {
			//	t.Errorf("normalModel() = %v, want %v", rows, tt.want)
			//}
		})
	}
}

func createDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(""), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.Use(MegaGormAuditPlugin{})
	return db, nil
}

func deletedAtNull() gorm.DeletedAt {
	return gorm.DeletedAt{
		Time:  time.Time{},
		Valid: false,
	}
}
