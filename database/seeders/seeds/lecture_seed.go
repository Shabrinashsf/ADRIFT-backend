package seeds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"ADRIFT-backend/internal/entity"

	"gorm.io/gorm"
)

func ListLectureSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/json/lectures.json")
	if err != nil {
		return err
	}
	defer func() {
		if err := jsonFile.Close(); err != nil {
			log.Printf("failed to close lectures.json: %v", err)
		}
	}()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var listLecture []entity.Lecture
	if err := json.Unmarshal(jsonData, &listLecture); err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entity.Lecture{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entity.Lecture{}); err != nil {
			return err
		}
	}

	inserted := 0
	skipped := 0
	skippedDetails := make([]string, 0, 16)

	for _, data := range listLecture {
		if strings.TrimSpace(data.Code) == "" {
			skipped++
			skippedDetails = append(skippedDetails, fmtSkippedLecture(data, "empty code"))
			continue
		}

		var lecture entity.Lecture
		err := db.Where(&entity.Lecture{Code: data.Code}).First(&lecture).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		isData := db.Find(&lecture, "code = ?", data.Code).RowsAffected
		if isData == 0 {
			if err := db.Create(&data).Error; err != nil {
				return err
			}
			inserted++
		} else {
			skipped++
			skippedDetails = append(skippedDetails, fmtSkippedLecture(data, "duplicate code"))
		}
	}

	log.Printf("lectures seed summary: total=%d inserted=%d skipped=%d", len(listLecture), inserted, skipped)
	if len(skippedDetails) > 0 {
		log.Printf("lectures skipped details: %s", strings.Join(skippedDetails, "; "))
	}

	return nil
}

func fmtSkippedLecture(data entity.Lecture, reason string) string {
	return fmt.Sprintf("id=%s code=%s name=%s reason=%s", data.ID.String(), data.Code, data.Name, reason)
}
