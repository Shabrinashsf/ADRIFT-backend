package migrations

import (
	"ADRIFT-backend/internal/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	if err := db.AutoMigrate(
		&entity.User{},
		&entity.Course{},
		&entity.Prerequisite{},
		&entity.Lecture{},
		&entity.Schedule{},
		&entity.FRSPlan{},
		&entity.FRSPlanItem{},
		&entity.LabPath{},
		&entity.PathEdge{},
		&entity.StudentProgress{},
	); err != nil {
		return err
	}

	return nil
}
