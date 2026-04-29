package seeds

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ListPrerequisiteSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/json/prerequisites.json")
	if err != nil {
		return err
	}
	defer func() {
		if err := jsonFile.Close(); err != nil {
			log.Printf("failed to close prerequisites.json: %v", err)
		}
	}()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var listPrerequisite []entity.Prerequisite
	if err := json.Unmarshal(jsonData, &listPrerequisite); err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entity.Prerequisite{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entity.Prerequisite{}); err != nil {
			return err
		}
	}

	inserted := 0
	skipped := 0

	for _, data := range listPrerequisite {
		if data.ID == uuid.Nil || data.CourseID == uuid.Nil || data.RequireID == uuid.Nil {
			skipped++
			log.Printf("prerequisite seed skipped: empty id/course_id/require_id")
			continue
		}

		var prerequisite entity.Prerequisite
		err := db.Where(&entity.Prerequisite{CourseID: data.CourseID, RequireID: data.RequireID}).First(&prerequisite).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		isData := db.Find(&prerequisite, "course_id = ? AND require_id = ?", data.CourseID, data.RequireID).RowsAffected
		if isData == 0 {
			if err := db.Create(&data).Error; err != nil {
				return err
			}
			inserted++
		} else {
			skipped++
		}
	}

	log.Printf("prerequisites seed summary: total=%d inserted=%d skipped=%d", len(listPrerequisite), inserted, skipped)

	return nil
}
