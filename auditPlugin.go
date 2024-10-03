package MegaGormAudit

import (
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"reflect"
)

type MegaGormAuditPlugin struct{}

func (a MegaGormAuditPlugin) Name() string {
	return "MegaGormAuditPlugin"
}

func (a MegaGormAuditPlugin) Initialize(db *gorm.DB) error {
	err := db.Callback().Update().Replace("gorm:update", deleteAndCreate)
	return err
}

func deleteAndCreate(db *gorm.DB) {
	auditableModelReflect := reflect.ValueOf(db.Statement.Model).Elem().FieldByName("AuditableModel")

	if auditableModelReflect.IsValid() {
		auditableModel := auditableModelReflect.Interface().(AuditableModel)

		err := db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Delete(db.Statement.Model).Error; err != nil {
				db.AddError(err)
				return err
			}

			db.Statement.SetColumn("id", 0)
			if auditableModel.AuditParentID != nil {
				db.Statement.SetColumn("audit_parent_id", auditableModel.AuditParentID)
			} else {
				db.Statement.SetColumn("audit_parent_id", auditableModel.ID)
			}
			db.Statement.SetColumn("deleted_at", nil)
			db.Statement.SetColumn("last_changed_user", auditableModel.LastChangedUser)

			if err := tx.Create(db.Statement.Model).Error; err != nil {
				db.AddError(err)
				return err
			}

			return nil
		})

		if err != nil {
			db.AddError(err)
			return
		}

	} else {
		callbacks.Update(&callbacks.Config{})(db)
		return
	}
}

func (u *AuditableModel) BeforeDelete(tx *gorm.DB) (err error) {

	curTime := tx.Statement.DB.NowFunc()
	nano := curTime.UnixMilli()

	tx.Statement.AddClause(clause.Update{})
	tx.Statement.AddClause(clause.Set{
		{Column: clause.Column{Name: "deleted_at"}, Value: nano},
		{Column: clause.Column{Name: "last_changed_user"}, Value: u.LastChangedUser},
	})

	tx.Statement.SetColumn("deleted_at", nano)
	tx.Statement.SetColumn("last_changed_user", u.LastChangedUser)

	tx.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: clause.PrimaryColumn, Value: u.ID},
		clause.Eq{Column: "deleted_at", Value: 0},
	}})

	tx.Statement.Build(
		clause.Update{}.Name(),
		clause.Set{}.Name(),
		clause.Where{}.Name(),
	)

	return
}
