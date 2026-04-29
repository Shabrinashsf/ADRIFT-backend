package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"ADRIFT-backend/database/mappers"
	"ADRIFT-backend/internal/api/repository"
	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/entity"
	myerror "ADRIFT-backend/internal/pkg/error"
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
	scheduleUploadMetaPath = "database/json/schedule_upload_meta.json"
	frsTempPrefix          = "tmp/frs/"
	frsPrefix              = "frs/"
	schedulePendingTTL     = 24 * time.Hour
)

type scheduleUploadMeta struct {
	ObjectKey    string    `json:"object_key"`
	AcademicYear string    `json:"academic_year"`
	Term         string    `json:"term"`
	CreatedAt    time.Time `json:"created_at"`
}

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

	if err := s.enforcePendingUpload(ctx); err != nil {
		return dto.FRSUploadResponse{}, err
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

	scheduleData, err := os.ReadFile(scheduleJSONPath)
	if err != nil {
		_ = s.storage.DeleteFile(objectKey)
		_ = cleanupScheduleFiles()
		return dto.FRSUploadResponse{}, err
	}

	var raw map[string]map[string][]mappers.ScheduleOutput
	if err := json.Unmarshal(scheduleData, &raw); err != nil {
		_ = s.storage.DeleteFile(objectKey)
		_ = cleanupScheduleFiles()
		return dto.FRSUploadResponse{}, err
	}

	lectures, err := s.frsRepo.ListLectures(ctx)
	if err != nil {
		_ = s.storage.DeleteFile(objectKey)
		_ = cleanupScheduleFiles()
		return dto.FRSUploadResponse{}, err
	}

	lectureByCode := buildLectureCodeMap(lectures)
	missingLectures := findMissingLectureCodes(raw, lectureByCode)

	publicLink := s.storage.GetPublicLink(objectKey)

	meta := scheduleUploadMeta{
		ObjectKey:    objectKey,
		AcademicYear: academicYear,
		Term:         string(term),
		CreatedAt:    time.Now(),
	}
	if err := writeScheduleUploadMeta(meta); err != nil {
		_ = s.storage.DeleteFile(objectKey)
		_ = cleanupScheduleFiles()
		return dto.FRSUploadResponse{}, err
	}

	return dto.FRSUploadResponse{
		FileURL:             publicLink,
		ObjectKey:           objectKey,
		FileName:            fileName,
		AcademicYear:        academicYear,
		Term:                string(term),
		NullRecords:         report.Records,
		MissingLectureCodes: formatMissingList(missingLectures),
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

	if err := removeFileIfExists(scheduleUploadMetaPath); err != nil {
		return err
	}

	return nil
}

func (s *frsService) SubmitSchedule(ctx context.Context, req dto.FRSSubmitRequest) (dto.FRSSubmitResponse, error) {
	objectKey, err := normalizeScheduleObjectKey(req.ObjectKey)
	if err != nil {
		return dto.FRSSubmitResponse{}, err
	}

	sourcePath := filepath.Join("./assets", filepath.FromSlash(objectKey))
	if !fileExists(sourcePath) {
		return dto.FRSSubmitResponse{}, dto.ErrScheduleFileNotFound
	}

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

	lectures, err := s.frsRepo.ListLectures(ctx)
	if err != nil {
		return dto.FRSSubmitResponse{}, err
	}

	lectureByCode := buildLectureCodeMap(lectures)

	var schedules []entity.Schedule
	for _, semesterMap := range raw {
		for _, items := range semesterMap {
			for _, item := range items {
				lectureKey := normalizeLectureCode(item.LectureCode)
				if lectureKey == "" {
					lectureKey = "SKPB"
				}

				lectureID, okLecture := lectureByCode[lectureKey]
				if !okLecture {
					var err error
					lectureID, err = s.frsRepo.EnsureLectureByCode(ctx, lectureKey)
					if err != nil {
						return dto.FRSSubmitResponse{}, err
					}
					lectureByCode[lectureKey] = lectureID
				}

				schedule := entity.Schedule{
					ID:           item.ID,
					CourseName:   strings.TrimSpace(item.CourseName),
					LectureID:    lectureID,
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

	if err := s.storage.DeleteFile(objectKey); err != nil && !errors.Is(err, os.ErrNotExist) {
		return dto.FRSSubmitResponse{}, err
	}
	if err := cleanupScheduleFiles(); err != nil {
		return dto.FRSSubmitResponse{}, err
	}
	if err := removeFileIfExists(scheduleUploadMetaPath); err != nil {
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

func buildLectureCodeMap(lectures []entity.Lecture) map[string]uuid.UUID {
	result := make(map[string]uuid.UUID, len(lectures))
	for _, lecture := range lectures {
		key := normalizeLectureCode(lecture.Code)
		if key == "" {
			continue
		}
		result[key] = lecture.ID
	}
	return result
}

func normalizeLectureCode(code string) string {
	cleaned := strings.NewReplacer(",", " ", ";", " ", "/", " ", "&", " ").Replace(code)
	cleaned = strings.ToUpper(strings.TrimSpace(cleaned))
	parts := strings.Fields(cleaned)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func buildMissingLectureMessage(missingLectures map[string]struct{}) string {
	return fmt.Sprintf("lecture_code tidak ditemukan: %s", formatMissingSet(missingLectures))
}

func findMissingLectureCodes(raw map[string]map[string][]mappers.ScheduleOutput, lectureByCode map[string]uuid.UUID) map[string]struct{} {
	missingLectures := make(map[string]struct{})

	for _, semesterMap := range raw {
		for _, items := range semesterMap {
			for _, item := range items {
				lectureKey := normalizeLectureCode(item.LectureCode)
				if lectureKey == "" {
					continue
				}
				if _, ok := lectureByCode[lectureKey]; !ok {
					missingLectures[strings.TrimSpace(item.LectureCode)] = struct{}{}
				}
			}
		}
	}

	return missingLectures
}

func formatMissingSet(values map[string]struct{}) string {
	list := make([]string, 0, len(values))
	for value := range values {
		cleaned := strings.TrimSpace(value)
		if cleaned == "" {
			cleaned = "<empty>"
		}
		list = append(list, cleaned)
	}
	sort.Strings(list)
	return strings.Join(list, ", ")
}

func formatMissingList(values map[string]struct{}) []string {
	list := make([]string, 0, len(values))
	for value := range values {
		cleaned := strings.TrimSpace(value)
		if cleaned == "" {
			cleaned = "<empty>"
		}
		list = append(list, cleaned)
	}
	sort.Strings(list)
	return list
}

func (s *frsService) enforcePendingUpload(ctx context.Context) error {
	_ = ctx
	meta, err := readScheduleUploadMeta()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if meta == nil {
		return nil
	}

	if time.Since(meta.CreatedAt) > schedulePendingTTL {
		return s.cleanupPendingUpload(meta)
	}

	return buildSchedulePendingError(meta)
}

func (s *frsService) cleanupPendingUpload(meta *scheduleUploadMeta) error {
	if meta == nil {
		return nil
	}
	if meta.ObjectKey != "" {
		if err := s.storage.DeleteFile(meta.ObjectKey); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := cleanupScheduleFiles(); err != nil {
		return err
	}
	return removeFileIfExists(scheduleUploadMetaPath)
}

func buildSchedulePendingError(meta *scheduleUploadMeta) error {
	message := "masih ada upload jadwal yang belum direvise atau disubmit"
	if meta != nil && strings.TrimSpace(meta.ObjectKey) != "" {
		message = fmt.Sprintf("%s (object_key: %s)", message, meta.ObjectKey)
	}
	return myerror.New(message, http.StatusConflict)
}

func readScheduleUploadMeta() (*scheduleUploadMeta, error) {
	data, err := os.ReadFile(scheduleUploadMetaPath)
	if err != nil {
		return nil, err
	}

	var meta scheduleUploadMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func writeScheduleUploadMeta(meta scheduleUploadMeta) error {
	payload, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(scheduleUploadMetaPath, payload, 0644)
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

func normalizeScheduleObjectKey(objectKey string) (string, error) {
	cleaned := strings.TrimSpace(objectKey)
	if cleaned == "" {
		return "", dto.ErrScheduleFileNotFound
	}

	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = filepath.ToSlash(filepath.Clean(cleaned))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "..") {
		return "", dto.ErrScheduleFileNotFound
	}
	if !strings.HasPrefix(cleaned, frsPrefix) {
		return "", dto.ErrScheduleFileNotFound
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
