package seeds

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"ADRIFT-backend/internal/entity"

	"gorm.io/gorm"
)

func ListCourseSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/json/courses.json")
	if err != nil {
		return err
	}
	defer func() {
		if err := jsonFile.Close(); err != nil {
			log.Printf("failed to close courses.json: %v", err)
		}
	}()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var listCourse []entity.Course
	if err := json.Unmarshal(jsonData, &listCourse); err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entity.Course{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entity.Course{}); err != nil {
			return err
		}
	}

	inserted := 0
	skipped := 0

	for _, data := range listCourse {
		if strings.TrimSpace(data.Code) == "" {
			skipped++
			log.Printf("course seed skipped: empty code")
			continue
		}

		var course entity.Course
		err := db.Where(&entity.Course{Code: data.Code}).First(&course).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		isData := db.Find(&course, "code = ?", data.Code).RowsAffected
		if isData == 0 {
			if err := db.Create(&data).Error; err != nil {
				return err
			}
			inserted++
		} else {
			skipped++
		}
	}

	log.Printf("courses seed summary: total=%d inserted=%d skipped=%d", len(listCourse), inserted, skipped)

	return nil
}
