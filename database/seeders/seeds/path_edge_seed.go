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

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ListPathEdgeSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/json/path_edges.json")
	if err != nil {
		return err
	}
	defer func() {
		if err := jsonFile.Close(); err != nil {
			log.Printf("failed to close path_edges.json: %v", err)
		}
	}()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var listPathEdge []entity.PathEdge
	if err := json.Unmarshal(jsonData, &listPathEdge); err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entity.PathEdge{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entity.PathEdge{}); err != nil {
			return err
		}
	}

	inserted := 0
	skipped := 0
	skippedDetails := make([]string, 0, 16)

	for _, data := range listPathEdge {
		if data.ID == uuid.Nil || data.FromCourseID == uuid.Nil || data.ToCourseID == uuid.Nil {
			skipped++
			skippedDetails = append(skippedDetails, fmtSkippedPathEdge(data, "empty id/from_course_id/to_course_id"))
			continue
		}

		var pathEdge entity.PathEdge
		err := db.Where(&entity.PathEdge{FromCourseID: data.FromCourseID, ToCourseID: data.ToCourseID}).First(&pathEdge).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		isData := db.Find(&pathEdge, "from_course_id = ? AND to_course_id = ?", data.FromCourseID, data.ToCourseID).RowsAffected
		if isData == 0 {
			if err := db.Create(&data).Error; err != nil {
				return err
			}
			inserted++
		} else {
			skipped++
			skippedDetails = append(skippedDetails, fmtSkippedPathEdge(data, "duplicate from/to"))
		}
	}

	log.Printf("path edges seed summary: total=%d inserted=%d skipped=%d", len(listPathEdge), inserted, skipped)
	if len(skippedDetails) > 0 {
		log.Printf("path edges skipped details: %s", strings.Join(skippedDetails, "; "))
	}

	return nil
}

func fmtSkippedPathEdge(data entity.PathEdge, reason string) string {
	return fmt.Sprintf("id=%s from=%s to=%s reason=%s", data.ID.String(), data.FromCourseID.String(), data.ToCourseID.String(), reason)
}
