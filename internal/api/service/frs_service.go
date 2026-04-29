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
	"ADRIFT-backend/internal/pkg/pagination"
	"ADRIFT-backend/internal/pkg/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	FRSService interface {
		UploadScheduleFile(ctx context.Context, req dto.FRSUploadRequest) (dto.FRSUploadResponse, error)
		DeleteScheduleArtifacts(ctx context.Context, req dto.FRSUploadDeleteRequest) error
		SubmitSchedule(ctx context.Context, req dto.FRSSubmitRequest) (dto.FRSSubmitResponse, error)
		ListSchedules(ctx context.Context, meta pagination.Meta) ([]dto.ScheduleResponse, pagination.Meta, error)
		CreateFRSPlan(ctx context.Context, userID string, req dto.CreateFRSPlanRequest) error
		ListFRSPlans(ctx context.Context, userID string) ([]dto.FRSPlanListItem, error)
		GetFRSPlanDetail(ctx context.Context, planID string, userID string) (dto.FRSPlanDetailResponse, error)
		FindAlternatives(ctx context.Context, req dto.AlternativeScheduleRequest) (dto.AlternativeScheduleResponse, error)
		DeleteFRSPlan(ctx context.Context, planID string, userID string) error
	}

	frsService struct {
		frsRepo repository.FRSRepository
		storage storage.FileSystemStorage
		db      *gorm.DB
	}
)

func NewFRSService(frsRepo repository.FRSRepository, storage storage.FileSystemStorage, db *gorm.DB) FRSService {
	return &frsService{
		frsRepo: frsRepo,
		storage: storage,
		db:      db,
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

var wibLocation, _ = time.LoadLocation("Asia/Jakarta")

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

func (s *frsService) ListSchedules(ctx context.Context, meta pagination.Meta) ([]dto.ScheduleResponse, pagination.Meta, error) {
	schedules, totalData, err := s.frsRepo.ListSchedules(ctx, nil, meta)
	if err != nil {
		return nil, meta, err
	}

	meta.Count(totalData)

	result := make([]dto.ScheduleResponse, 0, len(schedules))
	for _, schedule := range schedules {
		mappedSchedule := dto.ScheduleResponse{
			ID:          schedule.ID.String(),
			CourseName:  schedule.CourseName,
			SKS:         schedule.SKS,
			Class:       schedule.Class,
			Day:         schedule.Day,
			StartTime:   schedule.StartTime.In(wibLocation).Format("15:04"),
			EndTime:     schedule.EndTime.In(wibLocation).Format("15:04"),
			LectureID:   schedule.LectureID.String(),
			LectureName: schedule.Lecture.Name,
			Room:        schedule.Room,
			Capacity:    schedule.Capacity,
			Semester:    schedule.Semester,
			Prodi:       schedule.Prodi,
		}
		result = append(result, mappedSchedule)
	}

	return result, meta, nil
}

func (s *frsService) CreateFRSPlan(ctx context.Context, userID string, req dto.CreateFRSPlanRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return dto.ErrPlanNotOwnedByUser
	}

	academicYear, err := normalizeAcademicYear(req.AcademicYear)
	if err != nil {
		return err
	}

	term, err := normalizeTerm(req.Term)
	if err != nil {
		return err
	}

	scheduleUUIDs := make([]uuid.UUID, 0, len(req.ScheduleIDs))
	for _, idStr := range req.ScheduleIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return dto.ErrScheduleIDNotFound
		}
		scheduleUUIDs = append(scheduleUUIDs, id)
	}

	schedules, err := s.frsRepo.FindSchedulesByIDs(ctx, scheduleUUIDs)
	if err != nil {
		return err
	}
	if len(schedules) != len(scheduleUUIDs) {
		return dto.ErrScheduleIDNotFound
	}

	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	plan := entity.FRSPlan{
		UserID:       userUUID,
		PlanName:     req.PlanName,
		AcademicYear: academicYear,
		Term:         term,
		TotalCredit:  req.TotalCredit,
	}

	createdPlan, err := s.frsRepo.CreateFRSPlan(ctx, tx, plan)
	if err != nil {
		tx.Rollback()
		return err
	}

	items := make([]entity.FRSPlanItem, 0, len(scheduleUUIDs))
	for _, scheduleID := range scheduleUUIDs {
		items = append(items, entity.FRSPlanItem{
			FRSPlanID:  createdPlan.ID,
			ScheduleID: scheduleID,
		})
	}

	if err := s.frsRepo.CreateFRSPlanItems(ctx, tx, items); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *frsService) ListFRSPlans(ctx context.Context, userID string) ([]dto.FRSPlanListItem, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, dto.ErrPlanNotOwnedByUser
	}

	plans, err := s.frsRepo.ListFRSPlansByUserID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.FRSPlanListItem, 0, len(plans))
	for _, plan := range plans {
		result = append(result, dto.FRSPlanListItem{
			ID:           plan.ID.String(),
			PlanName:     plan.PlanName,
			AcademicYear: plan.AcademicYear,
			Term:         string(plan.Term),
			TotalCredit:  plan.TotalCredit,
			CourseCount:  len(plan.FRSPlanItems),
			CreatedAt:    plan.CreatedAt,
		})
	}

	return result, nil
}

func (s *frsService) GetFRSPlanDetail(ctx context.Context, planID string, userID string) (dto.FRSPlanDetailResponse, error) {
	planUUID, err := uuid.Parse(planID)
	if err != nil {
		return dto.FRSPlanDetailResponse{}, dto.ErrPlanNotFound
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return dto.FRSPlanDetailResponse{}, dto.ErrPlanNotOwnedByUser
	}

	plan, err := s.frsRepo.GetFRSPlanByID(ctx, planUUID, userUUID)
	if err != nil {
		return dto.FRSPlanDetailResponse{}, dto.ErrPlanNotFound
	}

	items := make([]dto.FRSPlanItemDetail, 0, len(plan.FRSPlanItems))
	for _, item := range plan.FRSPlanItems {
		schedule := item.Schedule
		itemDetail := dto.FRSPlanItemDetail{
			ID:          item.ID.String(),
			ScheduleID:  schedule.ID.String(),
			CourseName:  schedule.CourseName,
			Class:       schedule.Class,
			Day:         string(schedule.Day),
			StartTime:   schedule.StartTime.In(wibLocation).Format("15:04"),
			EndTime:     schedule.EndTime.In(wibLocation).Format("15:04"),
			LectureName: schedule.Lecture.Name,
			Room:        schedule.Room,
			Credit:      schedule.SKS,
		}
		items = append(items, itemDetail)
	}

	return dto.FRSPlanDetailResponse{
		ID:           plan.ID.String(),
		PlanName:     plan.PlanName,
		AcademicYear: plan.AcademicYear,
		Term:         string(plan.Term),
		TotalCredit:  plan.TotalCredit,
		Items:        items,
	}, nil
}

func (s *frsService) FindAlternatives(ctx context.Context, req dto.AlternativeScheduleRequest) (dto.AlternativeScheduleResponse, error) {
	academicYear, err := normalizeAcademicYear(req.AcademicYear)
	if err != nil {
		return dto.AlternativeScheduleResponse{}, err
	}

	term, err := normalizeTerm(req.Term)
	if err != nil {
		return dto.AlternativeScheduleResponse{}, err
	}

	scheduleUUIDs := make([]uuid.UUID, 0, len(req.ScheduleIDs))
	for _, idStr := range req.ScheduleIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return dto.AlternativeScheduleResponse{}, dto.ErrScheduleIDNotFound
		}
		scheduleUUIDs = append(scheduleUUIDs, id)
	}

	selectedSchedules, err := s.frsRepo.FindSchedulesByIDs(ctx, scheduleUUIDs)
	if err != nil {
		return dto.AlternativeScheduleResponse{}, err
	}
	if len(selectedSchedules) != len(scheduleUUIDs) {
		return dto.AlternativeScheduleResponse{}, dto.ErrScheduleIDNotFound
	}

	selectedMap := make(map[uuid.UUID]entity.Schedule, len(selectedSchedules))
	for _, sch := range selectedSchedules {
		selectedMap[sch.ID] = sch
	}

	lectureIDs := make([]uuid.UUID, 0)
	lectureSeen := make(map[uuid.UUID]bool)
	courseNames := make([]string, 0)
	courseSeen := make(map[string]bool)
	for _, sch := range selectedSchedules {
		if !lectureSeen[sch.LectureID] {
			lectureSeen[sch.LectureID] = true
			lectureIDs = append(lectureIDs, sch.LectureID)
		}
		if !courseSeen[sch.CourseName] {
			courseSeen[sch.CourseName] = true
			courseNames = append(courseNames, sch.CourseName)
		}
	}

	var alternatives []dto.AlternativeGroup

	lecturerAlts := s.buildLecturerPriorityAlternatives(ctx, selectedSchedules, lectureIDs, academicYear, term, selectedMap)
	alternatives = append(alternatives, lecturerAlts...)

	courseAlts := s.buildCoursePriorityAlternatives(ctx, selectedSchedules, courseNames, academicYear, term, selectedMap)
	alternatives = append(alternatives, courseAlts...)

	if len(alternatives) == 0 {
		return dto.AlternativeScheduleResponse{}, dto.ErrNoAlternativeFound
	}

	return dto.AlternativeScheduleResponse{
		Alternatives: alternatives,
	}, nil
}

func (s *frsService) buildLecturerPriorityAlternatives(
	ctx context.Context,
	selectedSchedules []entity.Schedule,
	lectureIDs []uuid.UUID,
	academicYear string,
	term entity.TermSemester,
	selectedMap map[uuid.UUID]entity.Schedule,
) []dto.AlternativeGroup {
	lecturerSchedules, err := s.frsRepo.FindSchedulesByLectureIDsAndTerm(ctx, lectureIDs, academicYear, term)
	if err != nil {
		return nil
	}

	courseToAlternatives := buildCourseAlternativesMap(selectedSchedules, lecturerSchedules, selectedMap)
	selectedByCourse := make(map[string]int)
	for i, sch := range selectedSchedules {
		selectedByCourse[sch.CourseName] = i
	}

	return generateAlternativeGroups(selectedSchedules, courseToAlternatives, selectedByCourse, "Mendahulukan dosen yang sama", 5)
}

func (s *frsService) buildCoursePriorityAlternatives(
	ctx context.Context,
	selectedSchedules []entity.Schedule,
	courseNames []string,
	academicYear string,
	term entity.TermSemester,
	selectedMap map[uuid.UUID]entity.Schedule,
) []dto.AlternativeGroup {
	courseSchedules, err := s.frsRepo.FindSchedulesByCourseNamesAndTerm(ctx, courseNames, academicYear, term)
	if err != nil {
		return nil
	}

	courseToAlternatives := buildCourseAlternativesMap(selectedSchedules, courseSchedules, selectedMap)
	selectedByCourse := make(map[string]int)
	for i, sch := range selectedSchedules {
		selectedByCourse[sch.CourseName] = i
	}

	return generateAlternativeGroups(selectedSchedules, courseToAlternatives, selectedByCourse, "Mendahulukan matkul yang sama", 5)
}

func buildCourseAlternativesMap(
	selectedSchedules []entity.Schedule,
	candidateSchedules []entity.Schedule,
	selectedMap map[uuid.UUID]entity.Schedule,
) map[int][]entity.Schedule {
	selectedByCourse := make(map[string]int)
	courseOrder := 0
	for _, sch := range selectedSchedules {
		if _, exists := selectedByCourse[sch.CourseName]; !exists {
			selectedByCourse[sch.CourseName] = courseOrder
			courseOrder++
		}
	}

	alternatives := make(map[int][]entity.Schedule)
	for _, sch := range candidateSchedules {
		if _, isSelected := selectedMap[sch.ID]; isSelected {
			continue
		}
		idx, ok := selectedByCourse[sch.CourseName]
		if !ok {
			continue
		}
		alternatives[idx] = append(alternatives[idx], sch)
	}

	return alternatives
}

type courseSlot struct {
	courseIdx  int
	candidates []entity.Schedule
}

func generateAlternativeGroups(
	selectedSchedules []entity.Schedule,
	courseToAlternatives map[int][]entity.Schedule,
	selectedByCourse map[string]int,
	priorityNote string,
	maxGroups int,
) []dto.AlternativeGroup {
	courseByCourseName := make(map[string]entity.Schedule)
	for _, sch := range selectedSchedules {
		courseByCourseName[sch.CourseName] = sch
	}

	courseOrder := make([]string, 0, len(selectedByCourse))
	for name := range selectedByCourse {
		courseOrder = append(courseOrder, name)
	}
	sort.Strings(courseOrder)

	slots := make([]courseSlot, 0, len(courseOrder))
	for _, name := range courseOrder {
		idx := selectedByCourse[name]
		alts, hasAlts := courseToAlternatives[idx]
		if hasAlts && len(alts) > 0 {
			slots = append(slots, courseSlot{
				courseIdx:  idx,
				candidates: alts,
			})
		} else {
			original := courseByCourseName[name]
			slots = append(slots, courseSlot{
				courseIdx:  idx,
				candidates: []entity.Schedule{original},
			})
		}
	}

	if len(slots) == 0 {
		return nil
	}

	var results []dto.AlternativeGroup

	var generate func(slotIdx int, current []dto.AlternativeScheduleItem)
	generate = func(slotIdx int, current []dto.AlternativeScheduleItem) {
		if len(results) >= maxGroups {
			return
		}
		if slotIdx >= len(slots) {
			if !hasTimeConflict(current) {
				group := dto.AlternativeGroup{
					PriorityNote: priorityNote,
					Schedules:    make([]dto.AlternativeScheduleItem, len(current)),
				}
				copy(group.Schedules, current)
				results = append(results, group)
			}
			return
		}

		for _, candidate := range slots[slotIdx].candidates {
			item := scheduleToAlternativeItem(candidate)
			current = append(current, item)
			generate(slotIdx+1, current)
			current = current[:len(current)-1]
			if len(results) >= maxGroups {
				return
			}
		}
	}

	generate(0, nil)
	return results
}

func scheduleToAlternativeItem(s entity.Schedule) dto.AlternativeScheduleItem {
	lectureName := ""
	if s.Lecture != nil {
		lectureName = s.Lecture.Name
	}
	return dto.AlternativeScheduleItem{
		ScheduleID:  s.ID.String(),
		CourseName:  s.CourseName,
		Class:       s.Class,
		LectureName: lectureName,
		Day:         string(s.Day),
StartAt:     s.StartTime.In(wibLocation).Format("15:04"),
	EndAt:       s.EndTime.In(wibLocation).Format("15:04"),
		SKS:         s.SKS,
	}
}

func hasTimeConflict(items []dto.AlternativeScheduleItem) bool {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Day == items[j].Day {
				if items[i].StartAt < items[j].EndAt && items[j].StartAt < items[i].EndAt {
					return true
				}
			}
		}
	}
	return false
}

func (s *frsService) DeleteFRSPlan(ctx context.Context, planID string, userID string) error {
	planUUID, err := uuid.Parse(planID)
	if err != nil {
		return dto.ErrPlanNotFound
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return dto.ErrPlanNotOwnedByUser
	}

	if err := s.frsRepo.DeleteFRSPlan(ctx, planUUID, userUUID); err != nil {
		return dto.ErrPlanNotFound
	}

	return nil
}
