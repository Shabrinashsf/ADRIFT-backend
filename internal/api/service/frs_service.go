package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"ADRIFT-backend/database/mappers"
	"ADRIFT-backend/internal/api/repository"
	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/entity"
	"ADRIFT-backend/internal/pkg/storage"

	"github.com/google/uuid"
)

type (
	FRSService interface {
		UploadScheduleFile(ctx context.Context, req dto.FRSUploadRequest) (dto.FRSUploadResponse, error)
		DeleteScheduleArtifacts(ctx context.Context, req dto.FRSUploadDeleteRequest) error
		SubmitSchedule(ctx context.Context, req dto.FRSSubmitRequest) (dto.FRSSubmitResponse, error)
	}

	frsService struct {
		frsRepo repository.FRSRepository
		storage storage.FileSystemStorage
	}
)

func NewFRSService(frsRepo repository.FRSRepository, storage storage.FileSystemStorage) FRSService {
	return &frsService{
		frsRepo: frsRepo,
		storage: storage,
	}
}

const (
	scheduleJSONPath       = "database/json/schedule.json"
	scheduleReportJSONPath = "database/json/schedule_null_report.json"
	frsTempPrefix          = "tmp/frs/"
)

func (s *frsService) UploadScheduleFile(ctx context.Context, req dto.FRSUploadRequest) (dto.FRSUploadResponse, error) {
	academicYear, err := normalizeAcademicYear(req.AcademicYear)
	if err != nil {
		return dto.FRSUploadResponse{}, err
	}

	term, err := normalizeTerm(req.Term)
	if err != nil {
		return dto.FRSUploadResponse{}, err
	}

	exists, err := s.frsRepo.ScheduleExists(ctx, academicYear, term)
	if err != nil {
		return dto.FRSUploadResponse{}, err
	}
	if exists {
		return dto.FRSUploadResponse{}, dto.ErrScheduleAlreadyExist
	}

	tempObjectKey, err := normalizeTempObjectKey(req.ObjectKey)
	if err != nil {
		return dto.FRSUploadResponse{}, err
	}

	fileExt := strings.ToLower(filepath.Ext(tempObjectKey))
	if fileExt != ".xlsx" && fileExt != ".xls" {
		return dto.FRSUploadResponse{}, dto.ErrInvalidExcelFile
	}

	srcPath := filepath.Join("./assets", filepath.FromSlash(tempObjectKey))
	if !fileExists(srcPath) {
		return dto.FRSUploadResponse{}, dto.ErrTempFileNotFound
	}

	safeYear := strings.ReplaceAll(academicYear, "/", "-")
	fileName := fmt.Sprintf("frs_%s_%s_%s%s", safeYear, strings.ToLower(string(term)), uuid.New().String(), fileExt)
	objectKey := filepath.ToSlash(filepath.Join("frs", fileName))
	destPath := filepath.Join("./assets", filepath.FromSlash(objectKey))

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return dto.FRSUploadResponse{}, err
	}

	if err := moveFile(srcPath, destPath); err != nil {
		return dto.FRSUploadResponse{}, err
	}

	report, err := mappers.GenerateScheduleFiles(destPath, scheduleJSONPath, scheduleReportJSONPath, academicYear, term)
	if err != nil {
		_ = s.storage.DeleteFile(objectKey)
		_ = cleanupScheduleFiles()
		return dto.FRSUploadResponse{}, err
	}

	publicLink := s.storage.GetPublicLink(objectKey)

	return dto.FRSUploadResponse{
		FileURL:      publicLink,
		ObjectKey:    objectKey,
		FileName:     fileName,
		AcademicYear: academicYear,
		Term:         string(term),
		NullRecords:  report.Records,
	}, nil
}

func (s *frsService) DeleteScheduleArtifacts(ctx context.Context, req dto.FRSUploadDeleteRequest) error {
	_ = ctx

	if err := s.storage.DeleteFile(req.ObjectKey); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := cleanupScheduleFiles(); err != nil {
		return err
	}

	return nil
}

func (s *frsService) SubmitSchedule(ctx context.Context, req dto.FRSSubmitRequest) (dto.FRSSubmitResponse, error) {
	academicYear, err := normalizeAcademicYear(req.AcademicYear)
	if err != nil {
		return dto.FRSSubmitResponse{}, err
	}

	term, err := normalizeTerm(req.Term)
	if err != nil {
		return dto.FRSSubmitResponse{}, err
	}

	exists, err := s.frsRepo.ScheduleExists(ctx, academicYear, term)
	if err != nil {
		return dto.FRSSubmitResponse{}, err
	}
	if exists {
		return dto.FRSSubmitResponse{}, dto.ErrScheduleAlreadyExist
	}

	data, err := os.ReadFile(scheduleJSONPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return dto.FRSSubmitResponse{}, dto.ErrScheduleFileNotFound
		}
		return dto.FRSSubmitResponse{}, err
	}

	var raw map[string]map[string][]mappers.ScheduleOutput
	if err := json.Unmarshal(data, &raw); err != nil {
		return dto.FRSSubmitResponse{}, err
	}

	var schedules []entity.Schedule
	for _, semesterMap := range raw {
		for _, items := range semesterMap {
			for _, item := range items {
				schedule := entity.Schedule{
					ID:           item.ID,
					CourseID:     uuid.Nil, // TODO: resolve course_id from course name/code
					LectureID:    uuid.Nil, // TODO: resolve lecture_id from lecture code
					Class:        item.Class,
					Day:          item.Day,
					StartTime:    item.StartTime,
					EndTime:      item.EndTime,
					Room:         item.Room,
					Semester:     item.Semester,
					AcademicYear: item.AcademicYear,
					Capacity:     item.Capacity,
					SKS:          item.SKS,
					Prodi:        item.Prodi,
					Term:         item.Term,
				}
				schedules = append(schedules, schedule)
			}
		}
	}

	if err := s.frsRepo.CreateSchedules(ctx, schedules); err != nil {
		return dto.FRSSubmitResponse{}, err
	}

	return dto.FRSSubmitResponse{Inserted: len(schedules)}, nil
}

func normalizeAcademicYear(value string) (string, error) {
	cleaned := strings.TrimSpace(value)
	if cleaned == "" {
		return "", dto.ErrInvalidAcademicYear
	}

	cleaned = strings.ReplaceAll(cleaned, "-", "/")
	pattern := regexp.MustCompile(`^\d{4}/\d{4}$`)
	if !pattern.MatchString(cleaned) {
		return "", dto.ErrInvalidAcademicYear
	}

	parts := strings.Split(cleaned, "/")
	if len(parts) != 2 {
		return "", dto.ErrInvalidAcademicYear
	}

	startYear, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", dto.ErrInvalidAcademicYear
	}
	endYear, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", dto.ErrInvalidAcademicYear
	}
	if endYear != startYear+1 {
		return "", dto.ErrInvalidAcademicYear
	}

	return cleaned, nil
}

func normalizeTerm(value string) (entity.TermSemester, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "GANJIL" {
		return entity.TermSemesterGanjil, nil
	}
	if normalized == "GENAP" {
		return entity.TermSemesterGenap, nil
	}
	return "", dto.ErrInvalidTerm
}

func cleanupScheduleFiles() error {
	if err := removeFileIfExists(scheduleJSONPath); err != nil {
		return err
	}
	if err := removeFileIfExists(scheduleReportJSONPath); err != nil {
		return err
	}
	return nil
}

func normalizeTempObjectKey(objectKey string) (string, error) {
	cleaned := strings.TrimSpace(objectKey)
	if cleaned == "" {
		return "", dto.ErrTempFileNotFound
	}

	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = filepath.ToSlash(filepath.Clean(cleaned))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "..") {
		return "", dto.ErrTempFileNotFound
	}
	if !strings.HasPrefix(cleaned, frsTempPrefix) {
		return "", dto.ErrTempFileNotFound
	}

	return cleaned, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func moveFile(srcPath, destPath string) error {
	if err := os.Rename(srcPath, destPath); err == nil {
		return nil
	}

	return copyAndRemove(srcPath, destPath)
}

func copyAndRemove(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return os.Remove(srcPath)
}

func removeFileIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
