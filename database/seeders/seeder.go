package seeders

import (
	"ADRIFT-backend/database/seeders/seeds"

	"gorm.io/gorm"
)

func Seeder(db *gorm.DB) error {
	if err := seeds.ListUserSeeder(db); err != nil {
		return err
	}

	if err := seeds.ListCourseSeeder(db); err != nil {
		return err
	}

	if err := seeds.ListPrerequisiteSeeder(db); err != nil {
		return err
	}

	return nil
}
