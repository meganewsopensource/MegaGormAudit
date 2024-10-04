package MegaGormAudit

import (
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/soft_delete"
	"reflect"
	"strings"
	"testing"
	"time"
)

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

				return len(rows) > 0
			},
			afterCreate: nil,
			wantErr:     false,
		},
		{
			name: "Success, Update",
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

				return rows[0].Name == "teste atualizado"
			},
			wantErr: false,
		},
		{
			name: "Success, Delete",
			model: &Player{
				Model:    gorm.Model{},
				Name:     "teste",
				NickName: "teste",
			},
			want: []Player{},
			afterCreate: func(db *gorm.DB, model *Player) error {
				return db.Delete(model).Error
			},
			success: func(db *gorm.DB) bool {
				var rows []Player
				db.Find(&rows)

				if len(rows) == 0 {
					db.Unscoped().Find(&rows)
					return len(rows) == 1 && rows[0].DeletedAt.Valid
				}
				return false
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

			} else {

				var rows []Player
				db.Find(&rows)

				if !compareFunc(rows, tt.want) {
					t.Errorf("[]Player() = %v, want %v", rows, tt.want)
				}
			}
		})
	}
}

func TestAuditPlugin_AuditableModel(t *testing.T) {

	type Player struct {
		AuditableModel
		Name     string
		NickName string
	}

	type PlayerWithCorrectUniqueIndex struct {
		AuditableModel
		Name      string `gorm:"uniqueIndex:udx_name"`
		NickName  string
		DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:udx_name"`
	}
	type PlayerWithWrongUniqueIndex struct {
		AuditableModel
		Name     string `gorm:"uniqueIndex"`
		NickName string
	}
	type PlayerWithWrongUniqueIndex2 struct {
		AuditableModel
		Name      string `gorm:"uniqueIndex"`
		NickName  string
		DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex"`
	}

	type PlayerWithCorrectMultipleUniqueIndex struct {
		AuditableModel
		Name      string `gorm:"uniqueIndex:udx_name"`
		NickName  string `gorm:"uniqueIndex:udx_name"`
		Address   string
		DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:udx_name"`
	}
	type PlayerWithWrongMultipleUniqueIndex struct {
		AuditableModel
		Name      string `gorm:"uniqueIndex"`
		NickName  string `gorm:"uniqueIndex"`
		Address   string
		DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex"`
	}

	tests := []struct {
		name            string
		model           interface{}
		modelsToMigrate []interface{}
		create          func(db *gorm.DB) error
		afterCreate     func(db *gorm.DB, model interface{}) error
		successTest     func(db *gorm.DB, t *testing.T) bool
		wantErr         bool
	}{
		{
			name: "Success, Create",
			model: &Player{
				AuditableModel: AuditableModel{},
				Name:           "teste",
				NickName:       "teste",
			},
			modelsToMigrate: []interface{}{Player{}},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []Player
				db.Find(&rows)

				return len(rows) > 0
			},
			afterCreate: nil,
			wantErr:     false,
		},
		{
			name: "Success, Update",
			model: &Player{
				AuditableModel: AuditableModel{},
				Name:           "teste",
				NickName:       "teste",
			},
			modelsToMigrate: []interface{}{Player{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				model.(*Player).Name = "teste atualizado"
				return db.Updates(model).Error
			},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []Player
				db.Find(&rows)

				return rows[0].Name == "teste atualizado"
			},
			wantErr: false,
		},
		{
			name:    "Update failed on delete data",
			wantErr: true,
			model: &PlayerErrorOnDelete{
				AuditableModel: AuditableModel{},
				Name:           "teste",
				NickName:       "teste",
			},
			modelsToMigrate: []interface{}{PlayerErrorOnDelete{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {

				model.(*PlayerErrorOnDelete).Name = "teste atualizado"
				return db.Updates(model).Error
			},
			successTest: nil,
		},
		{
			name:    "Update failed while creating new data",
			wantErr: true,
			model: &PlayerErrorOnCreate{
				AuditableModel: AuditableModel{},
				Name:           "teste",
				NickName:       "teste",
			},
			modelsToMigrate: []interface{}{PlayerErrorOnCreate{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {

				model.(*PlayerErrorOnCreate).Name = "teste atualizado"
				return db.Updates(model).Error
			},
			successTest: nil,
		},
		{
			name: "Success on Delete",
			model: &Player{
				AuditableModel: AuditableModel{},
				Name:           "teste",
				NickName:       "teste",
			},
			modelsToMigrate: []interface{}{Player{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				return db.Delete(model).Error
			},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []Player
				db.Find(&rows)

				if len(rows) == 0 {
					db.Unscoped().Find(&rows)
					return len(rows) == 1 && rows[0].DeletedAt > 0
				}
				return false
			},
			wantErr: false,
		},
		{
			name:    "Success on parent ID",
			wantErr: false,
			model: &Player{
				AuditableModel: AuditableModel{},
				Name:           "teste",
				NickName:       "teste",
			},
			modelsToMigrate: []interface{}{Player{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {

				var rows []Player
				err := db.Unscoped().Find(&rows).Error
				if err == nil && len(rows) == 1 {

					rows[0].Name = "updated row"
					err = db.Updates(&rows[0]).Error
					if err != nil {
						return err
					}

				}
				return err
			},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []Player
				err := db.Unscoped().Find(&rows).Error
				if err == nil && len(rows) == 2 {

					if rows[0].DeletedAt != 0 && rows[0].AuditParentID == nil &&
						rows[1].ID == 2 && rows[1].DeletedAt == 0 && (rows[1].AuditParentID != nil && *rows[1].AuditParentID == 1) {
						return true
					} else {
						t.Errorf("failed on return sucess test")
					}
				} else {
					t.Errorf("failed on return sucess test")
				}
				return false
			},
		},
		{
			name:    "Success on multiples parent ID",
			wantErr: false,
			model: &Player{
				AuditableModel: AuditableModel{},
				Name:           "teste",
				NickName:       "teste",
			},
			modelsToMigrate: []interface{}{Player{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				model.(*Player).Name = "updated row"
				err := db.Updates(model).Error
				if err != nil {
					return err
				}

				model.(*Player).Name = "one more time updated row"
				err = db.Updates(model).Error
				if err != nil {
					return err
				}

				return err
			},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []Player
				err := db.Unscoped().Find(&rows).Error
				if err == nil && len(rows) == 3 {

					if rows[0].DeletedAt != 0 && rows[0].AuditParentID == nil &&
						rows[1].ID == 2 && rows[1].DeletedAt > 0 && (rows[1].AuditParentID != nil && *rows[1].AuditParentID == 1) &&
						rows[2].ID == 3 && rows[2].DeletedAt == 0 && (rows[2].AuditParentID != nil && *rows[2].AuditParentID == 1) {
						return true
					} else {
						t.Errorf("failed on return sucess test")
					}
				} else {
					t.Errorf("failed on return sucess test")
				}
				return false
			},
		},
		{
			name:    "Success on index unique",
			wantErr: false,
			model: &PlayerWithCorrectUniqueIndex{
				AuditableModel: AuditableModel{},
				Name:           "name test",
				NickName:       "nick test",
			},
			modelsToMigrate: []interface{}{PlayerWithCorrectUniqueIndex{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				model.(*PlayerWithCorrectUniqueIndex).NickName = "altered nick"
				err := db.Updates(model).Error
				if err != nil {
					return err
				}
				return err
			},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []PlayerWithCorrectUniqueIndex
				err := db.Unscoped().Find(&rows).Error
				if err == nil && len(rows) == 2 {
					if rows[0].Name == "name test" && rows[0].DeletedAt > 0 &&
						rows[1].ID == 2 && rows[1].Name == "name test" && rows[1].NickName == "altered nick" && rows[1].DeletedAt == 0 && (rows[1].AuditParentID != nil && *rows[1].AuditParentID == 1) {
						return true
					} else {
						t.Errorf("failed on return sucess test")
					}

				} else {
					t.Errorf("failed on return sucess test")
				}
				return false
			},
		},
		{
			name:    "Fail on unique index without deletedAt unique index",
			wantErr: true,
			model: &PlayerWithWrongUniqueIndex{
				AuditableModel: AuditableModel{},
				Name:           "name test",
				NickName:       "nick test",
			},
			modelsToMigrate: []interface{}{PlayerWithWrongUniqueIndex{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				model.(*PlayerWithWrongUniqueIndex).NickName = "altered nick"
				err := db.Updates(model).Error

				isUniqueIndexError := strings.Contains(err.Error(), "UNIQUE constraint failed")

				if err != nil && isUniqueIndexError {
					return err
				}
				return nil
			},
			successTest: nil,
		},
		{
			name:    "Fail on unique index with DeletedAt unique index without index name",
			wantErr: true,
			model: &PlayerWithWrongUniqueIndex2{
				AuditableModel: AuditableModel{},
				Name:           "name test",
				NickName:       "nick test",
			},
			modelsToMigrate: []interface{}{PlayerWithWrongUniqueIndex2{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				model.(*PlayerWithWrongUniqueIndex2).NickName = "altered nick"
				err := db.Updates(model).Error

				isUniqueIndexError := strings.Contains(err.Error(), "UNIQUE constraint failed")

				if err != nil && isUniqueIndexError {
					return err
				}
				return nil
			},
			successTest: nil,
		},

		{
			name:    "Success on multiples unique index",
			wantErr: false,
			model: &PlayerWithCorrectMultipleUniqueIndex{
				AuditableModel: AuditableModel{},
				Name:           "name test",
				NickName:       "nick test",
				Address:        "my  road street",
			},
			modelsToMigrate: []interface{}{PlayerWithCorrectMultipleUniqueIndex{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				model.(*PlayerWithCorrectMultipleUniqueIndex).Address = "big forest avenue"
				err := db.Updates(model).Error
				if err != nil {
					return err
				}
				return err
			},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []PlayerWithCorrectMultipleUniqueIndex
				err := db.Unscoped().Find(&rows).Error
				if err == nil && len(rows) == 2 {
					if rows[0].Name == "name test" && rows[0].DeletedAt > 0 &&
						rows[1].ID == 2 && rows[1].Name == "name test" && rows[1].NickName == "nick test" && rows[1].Address == "big forest avenue" && rows[1].DeletedAt == 0 && (rows[1].AuditParentID != nil && *rows[1].AuditParentID == 1) {
						return true
					} else {
						t.Errorf("failed on return sucess test")
					}

				} else {
					t.Errorf("failed on return sucess test")
				}
				return false
			},
		},

		{
			name:    "Failed on multiples unique index",
			wantErr: true,
			model: &PlayerWithWrongMultipleUniqueIndex{
				AuditableModel: AuditableModel{},
				Name:           "name test",
				NickName:       "nick test",
				Address:        "my  road street",
			},
			modelsToMigrate: []interface{}{PlayerWithWrongMultipleUniqueIndex{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				model.(*PlayerWithWrongMultipleUniqueIndex).Address = "big forest avenue"
				err := db.Updates(model).Error
				if err != nil {
					return err
				}
				return err
			},
			successTest: nil,
		},

		{
			name: "Success on update Last Changed User",
			model: &Player{
				AuditableModel: AuditableModel{LastChangedUser: "First User"},
				Name:           "name test",
				NickName:       "nick test",
			},
			modelsToMigrate: []interface{}{Player{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {
				model.(*Player).LastChangedUser = "Second User"
				model.(*Player).Name = "altered name"
				return db.Updates(model).Error
			},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []Player
				err := db.Unscoped().Find(&rows).Error
				if err != nil {
					t.Errorf("failed on return sucess test")
					return false
				}

				if rows != nil && len(rows) == 2 {
					if rows[0].Name == "name test" && rows[0].DeletedAt > 0 && rows[0].LastChangedUser == "Second User" &&
						rows[1].ID == 2 && rows[1].Name == "altered name" && rows[1].LastChangedUser == "Second User" && rows[1].DeletedAt == 0 && (rows[1].AuditParentID != nil && *rows[1].AuditParentID == 1) {
						return true
					} else {
						t.Errorf("failed on return sucess test")
					}
				}

				return rows[0].Name == "teste atualizado"
			},
			wantErr: false,
		},

		{
			name:    "Success with save point transaction",
			wantErr: false,
			model: &Player{
				AuditableModel: AuditableModel{LastChangedUser: "First User"},
				Name:           "name test",
				NickName:       "nick test",
			},
			modelsToMigrate: []interface{}{Player{}},
			afterCreate: func(db *gorm.DB, model interface{}) error {

				tx := db.Begin()

				model.(*Player).Name = "altered name"
				model.(*Player).NickName = "altered nick"
				err := tx.Updates(model).Error
				if err != nil {
					t.Errorf("failed on return sucess test")
					return err
				}

				tx.SavePoint("sp1")

				model.(*Player).Name = "altered name 2"
				model.(*Player).NickName = "altered nick 2"
				err = tx.Updates(model).Error
				if err != nil {
					t.Errorf("failed on return sucess test")
					return err
				}

				err = tx.RollbackTo("sp1").Error
				if err != nil {
					t.Errorf("failed on return sucess test")
					return err
				}

				return tx.Commit().Error

			},
			successTest: func(db *gorm.DB, t *testing.T) bool {
				var rows []Player
				err := db.Unscoped().Find(&rows).Error
				if err != nil {
					t.Errorf("failed on return sucess test")
					return false
				}
				if len(rows) == 2 {
					return rows[0].Name == "name test" && rows[0].DeletedAt > 0 && rows[0].NickName == "nick test" &&
						rows[1].Name == "altered name" && rows[1].DeletedAt == 0 && rows[1].NickName == "altered nick"
				}
				return false
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			db, err := createDatabase()
			if err != nil {
				t.Errorf("createDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = db.AutoMigrate(tt.modelsToMigrate...)
			if err != nil && !tt.wantErr {
				t.Errorf("AutoMigrate() error = %v", err)
				return
			}

			err = db.Create(tt.model).Error
			if err != nil && !tt.wantErr {
				t.Errorf("Create() error = %v", err)
				return
			}

			if tt.afterCreate != nil {
				err = tt.afterCreate(db, tt.model)
				if err != nil && !tt.wantErr {
					t.Errorf("No Expected Error = %v", err)
					return
				}

			}

			if tt.successTest == nil && tt.wantErr == false {
				t.Errorf("SuccessTest not implemented for case")
			} else if tt.successTest != nil && (!tt.successTest(db, t) == tt.wantErr == false) {
				t.Errorf("[]Player() = false, want true")
			}

		})
	}
}

func createDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(""), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.Use(MegaGormAuditPlugin{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func deletedAtNull() gorm.DeletedAt {
	return gorm.DeletedAt{
		Time:  time.Time{},
		Valid: false,
	}
}

type PlayerErrorOnDelete struct {
	AuditableModel
	Name     string
	NickName string
}

func (u *PlayerErrorOnDelete) BeforeDelete(tx *gorm.DB) (err error) {

	tx.Statement.AddClause(clause.Update{})
	tx.Statement.AddClause(clause.Set{
		{Column: clause.Column{Name: "a"}, Value: ""},
	})

	tx.Statement.SetColumn("a", u.LastChangedUser)

	tx.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: clause.PrimaryColumn, Value: u.ID},
		clause.Eq{Column: "a", Value: 0},
	}})

	tx.Statement.Build(
		clause.Update{}.Name(),
		clause.Set{}.Name(),
		clause.Where{}.Name(),
	)

	return
}

type PlayerErrorOnCreate struct {
	AuditableModel
	Name     string
	NickName string
}

func (u *PlayerErrorOnCreate) BeforeCreate(tx *gorm.DB) (err error) {
	err = errors.New("error while before create")
	tx.AddError(err)
	return err
}
