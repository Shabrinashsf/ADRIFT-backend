package service

import (
	"context"
	"time"

	"ADRIFT-backend/internal/api/repository"
	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	SkillTreeService interface {
		GetGraph(ctx context.Context) (dto.GraphResponse, error)
		GetNodeDetail(ctx context.Context, courseID uuid.UUID) (dto.NodeDetailResponse, error)
		GetNodeChain(ctx context.Context, courseID uuid.UUID) (dto.NodeChainResponse, error)
		GetProgressGraph(ctx context.Context, userID uuid.UUID) (dto.ProgressGraphResponse, error)
		GetProgressSummary(ctx context.Context, userID uuid.UUID, enrollmentYear int) (dto.ProgressSummaryResponse, error)
		ClaimCourse(ctx context.Context, userID, courseID uuid.UUID, grade *string) (dto.ClaimCourseResponse, error)
		UnclaimCourse(ctx context.Context, userID, courseID uuid.UUID) (dto.UnclaimCourseResponse, error)
	}

	skillTreeService struct {
		skillTreeRepo repository.SkillTreeRepository
	}
)

func NewSkillTreeService(stRepo repository.SkillTreeRepository) SkillTreeService {
	return &skillTreeService{
		skillTreeRepo: stRepo,
	}
}

// =========== PUBLIC GRAPH ===========

func (s *skillTreeService) GetGraph(ctx context.Context) (dto.GraphResponse, error) {
	courses, err := s.skillTreeRepo.GetAllCourses(ctx)
	if err != nil {
		return dto.GraphResponse{}, err
	}
	prereqs, err := s.skillTreeRepo.GetAllPrerequisites(ctx)
	if err != nil {
		return dto.GraphResponse{}, err
	}
	pathEdges, err := s.skillTreeRepo.GetAllPathEdges(ctx)
	if err != nil {
		return dto.GraphResponse{}, err
	}

	nodes := make([]dto.GraphNodeResponse, 0, len(courses))
	for _, c := range courses {
		nodes = append(nodes, courseToGraphNode(c))
	}

	edges := buildEdges(prereqs, pathEdges)

	return dto.GraphResponse{Nodes: nodes, Edges: edges}, nil
}

// =========== NODE DETAIL ===========

func (s *skillTreeService) GetNodeDetail(ctx context.Context, courseID uuid.UUID) (dto.NodeDetailResponse, error) {
	course, err := s.skillTreeRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.NodeDetailResponse{}, dto.ErrCourseNotFound
		}
		return dto.NodeDetailResponse{}, err
	}

	// Prerequisites: courses that THIS course requires
	prereqs := make([]dto.NodeCourseRef, 0, len(course.PrerequisitesRequiredFor))
	for _, p := range course.PrerequisitesRequiredFor {
		if p.Require != nil {
			prereqs = append(prereqs, dto.NodeCourseRef{
				ID:       p.Require.ID.String(),
				Code:     p.Require.Code,
				Name:     p.Require.Name,
				Credit:   p.Require.Credit,
				Semester: p.Require.Semester,
			})
		}
	}

	// Unlocks: courses that this course enables
	unlockCourses, err := s.skillTreeRepo.GetUnlocksByCourseID(ctx, courseID)
	if err != nil {
		return dto.NodeDetailResponse{}, err
	}
	unlocks := make([]dto.NodeCourseRef, 0, len(unlockCourses))
	for _, u := range unlockCourses {
		unlocks = append(unlocks, dto.NodeCourseRef{
			ID:         u.ID.String(),
			Code:       u.Code,
			Name:       u.Name,
			Credit:     u.Credit,
			Semester:   u.Semester,
			IsElective: u.IsElective,
		})
	}

	// Schedules this semester
	schedules, err := s.skillTreeRepo.GetSchedulesByCourseID(ctx, courseID)
	if err != nil {
		return dto.NodeDetailResponse{}, err
	}
	scheduleDTOs := make([]dto.NodeScheduleResponse, 0, len(schedules))
	for _, sch := range schedules {
		lectureName := ""
		if sch.Lecture != nil {
			lectureName = sch.Lecture.Name
		}
		scheduleDTOs = append(scheduleDTOs, dto.NodeScheduleResponse{
			ID:          sch.ID.String(),
			Class:       sch.Class,
			Day:         string(sch.Day),
			StartTime:   sch.StartTime.In(wibLocation).Format("15:04"),
			EndTime:     sch.EndTime.In(wibLocation).Format("15:04"),
			LectureName: lectureName,
			Room:        sch.Room,
			Capacity:    sch.Capacity,
		})
	}

	labPaths := []dto.LabPathResponse{}
	if course.LabPath != nil {
		labPaths = append(labPaths, dto.LabPathResponse{
			ID:    course.LabPath.ID.String(),
			Name:  course.LabPath.Name,
			Color: course.LabPath.Color,
		})
	}

	return dto.NodeDetailResponse{
		ID:                    course.ID.String(),
		Code:                  course.Code,
		Name:                  course.Name,
		Credit:                course.Credit,
		Semester:              course.Semester,
		IsElective:            course.IsElective,
		Description:           course.Description,
		LabPaths:              labPaths,
		Prerequisites:         prereqs,
		Unlocks:               unlocks,
		SchedulesThisSemester: scheduleDTOs,
	}, nil
}

// =========== NODE CHAIN (BFS) ===========

func (s *skillTreeService) GetNodeChain(ctx context.Context, courseID uuid.UUID) (dto.NodeChainResponse, error) {
	prereqs, err := s.skillTreeRepo.GetAllPrerequisites(ctx)
	if err != nil {
		return dto.NodeChainResponse{}, err
	}
	pathEdges, err := s.skillTreeRepo.GetAllPathEdges(ctx)
	if err != nil {
		return dto.NodeChainResponse{}, err
	}

	// Build adjacency maps
	// upstream: courseID -> list of courses it requires (prereq + path edges incoming)
	// downstream: courseID -> list of courses it unlocks
	upstreamMap := map[uuid.UUID][]uuid.UUID{}
	downstreamMap := map[uuid.UUID][]uuid.UUID{}

	for _, p := range prereqs {
		// p.CourseID requires p.RequireID (RequireID is upstream of CourseID)
		upstreamMap[p.CourseID] = append(upstreamMap[p.CourseID], p.RequireID)
		downstreamMap[p.RequireID] = append(downstreamMap[p.RequireID], p.CourseID)
	}
	for _, e := range pathEdges {
		upstreamMap[e.ToCourseID] = append(upstreamMap[e.ToCourseID], e.FromCourseID)
		downstreamMap[e.FromCourseID] = append(downstreamMap[e.FromCourseID], e.ToCourseID)
	}

	upstream := bfsIDs(courseID, upstreamMap)
	downstream := bfsIDs(courseID, downstreamMap)

	upstreamStrs := make([]string, 0, len(upstream))
	for id := range upstream {
		upstreamStrs = append(upstreamStrs, id.String())
	}
	downstreamStrs := make([]string, 0, len(downstream))
	for id := range downstream {
		downstreamStrs = append(downstreamStrs, id.String())
	}

	return dto.NodeChainResponse{
		Upstream:   upstreamStrs,
		Downstream: downstreamStrs,
	}, nil
}

// =========== PROGRESS GRAPH ===========

func (s *skillTreeService) GetProgressGraph(ctx context.Context, userID uuid.UUID) (dto.ProgressGraphResponse, error) {
	courses, err := s.skillTreeRepo.GetAllCourses(ctx)
	if err != nil {
		return dto.ProgressGraphResponse{}, err
	}
	prereqs, err := s.skillTreeRepo.GetAllPrerequisites(ctx)
	if err != nil {
		return dto.ProgressGraphResponse{}, err
	}
	pathEdges, err := s.skillTreeRepo.GetAllPathEdges(ctx)
	if err != nil {
		return dto.ProgressGraphResponse{}, err
	}
	progressList, err := s.skillTreeRepo.GetProgressByUserID(ctx, userID)
	if err != nil {
		return dto.ProgressGraphResponse{}, err
	}

	statusMap, gradeMap, claimedAtMap := buildStatusMaps(progressList)
	nodeStatuses := computeNodeStatuses(courses, prereqs, pathEdges, statusMap)

	nodes := make([]dto.ProgressNodeResponse, 0, len(courses))
	for _, c := range courses {
		status := nodeStatuses[c.ID]
		labPaths := []dto.LabPathResponse{}
		if c.LabPath != nil {
			labPaths = append(labPaths, dto.LabPathResponse{
				ID:    c.LabPath.ID.String(),
				Name:  c.LabPath.Name,
				Color: c.LabPath.Color,
			})
		}
		nodes = append(nodes, dto.ProgressNodeResponse{
			ID:          c.ID.String(),
			Code:        c.Code,
			Name:        c.Name,
			Credit:      c.Credit,
			Semester:    c.Semester,
			IsElective:  c.IsElective,
			Description: c.Description,
			Status:      string(status),
			Grade:       gradeMap[c.ID],
			ClaimedAt:   claimedAtMap[c.ID],
			LabPaths:    labPaths,
		})
	}

	edges := buildEdges(prereqs, pathEdges)
	return dto.ProgressGraphResponse{Nodes: nodes, Edges: edges}, nil
}

// =========== PROGRESS SUMMARY ===========

func (s *skillTreeService) GetProgressSummary(ctx context.Context, userID uuid.UUID, enrollmentYear int) (dto.ProgressSummaryResponse, error) {
	// Fetch enrollment year from DB if not provided
	if enrollmentYear == 0 {
		user, err := s.skillTreeRepo.GetUserByID(ctx, userID)
		if err == nil {
			enrollmentYear = user.EnrollmentYear
		}
	}

	courses, err := s.skillTreeRepo.GetAllCourses(ctx)
	if err != nil {
		return dto.ProgressSummaryResponse{}, err
	}
	prereqs, err := s.skillTreeRepo.GetAllPrerequisites(ctx)
	if err != nil {
		return dto.ProgressSummaryResponse{}, err
	}
	pathEdges, err := s.skillTreeRepo.GetAllPathEdges(ctx)
	if err != nil {
		return dto.ProgressSummaryResponse{}, err
	}
	progressList, err := s.skillTreeRepo.GetProgressByUserID(ctx, userID)
	if err != nil {
		return dto.ProgressSummaryResponse{}, err
	}

	statusMap, _, _ := buildStatusMaps(progressList)
	nodeStatuses := computeNodeStatuses(courses, prereqs, pathEdges, statusMap)

	totalCourses := len(courses)
	completed, available, locked := 0, 0, 0
	totalCreditsCompleted := 0
	totalCreditsRequired := 0

	for _, c := range courses {
		totalCreditsRequired += c.Credit
		st := nodeStatuses[c.ID]
		switch st {
		case entity.NodeStatusCompleted: // COMPLETED
			completed++
			totalCreditsCompleted += c.Credit
		case entity.NodeStatusAvailable:
			available++
		default:
			locked++
		}
	}

	currentSemester := (time.Now().Year() - enrollmentYear) * 2
	if time.Now().Month() > 6 {
		currentSemester++
	}
	if currentSemester < 1 {
		currentSemester = 1
	}

	progressPct := 0.0
	if totalCreditsRequired > 0 {
		progressPct = float64(totalCreditsCompleted) / float64(totalCreditsRequired) * 100
	}

	return dto.ProgressSummaryResponse{
		TotalCourses:            totalCourses,
		Completed:               completed,
		Available:               available,
		Locked:                  locked,
		TotalCreditsCompleted:   totalCreditsCompleted,
		TotalCreditsRequired:    totalCreditsRequired,
		CurrentSemesterEstimate: currentSemester,
		EnrollmentYear:          enrollmentYear,
		ProgressPercentage:      progressPct,
	}, nil
}

// =========== CLAIM ===========

func (s *skillTreeService) ClaimCourse(ctx context.Context, userID, courseID uuid.UUID, grade *string) (dto.ClaimCourseResponse, error) {
	// Check course exists
	course, err := s.skillTreeRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		return dto.ClaimCourseResponse{}, dto.ErrCourseNotFound
	}

	// Compute current status
	courses, _ := s.skillTreeRepo.GetAllCourses(ctx)
	prereqs, _ := s.skillTreeRepo.GetAllPrerequisites(ctx)
	pathEdges, _ := s.skillTreeRepo.GetAllPathEdges(ctx)
	progressList, _ := s.skillTreeRepo.GetProgressByUserID(ctx, userID)
	statusMap, _, _ := buildStatusMaps(progressList)
	nodeStatuses := computeNodeStatuses(courses, prereqs, pathEdges, statusMap)

	currentStatus := nodeStatuses[courseID]
	if currentStatus == entity.NodeStatusLocked {
		return dto.ClaimCourseResponse{}, dto.ErrCourseNotAvailable
	}

	// Upsert progress as COMPLETED
	now := time.Now()
	progress := &entity.StudentProgress{
		ID:        uuid.New(),
		UserID:    userID,
		CourseID:  courseID,
		Status:    entity.NodeStatusCompleted,
		Grade:     grade,
		ClaimedAt: &now,
	}

	// If existing, update it
	existing, err := s.skillTreeRepo.GetProgressByCourseAndUser(ctx, userID, courseID)
	if err == nil && existing != nil {
		existing.Status = entity.NodeStatusCompleted
		existing.Grade = grade
		existing.ClaimedAt = &now
		progress = existing
	}

	if err := s.skillTreeRepo.UpsertProgress(ctx, progress); err != nil {
		return dto.ClaimCourseResponse{}, err
	}

	// Recompute and find newly AVAILABLE nodes
	_ = course
	progressList, _ = s.skillTreeRepo.GetProgressByUserID(ctx, userID)
	statusMap, _, _ = buildStatusMaps(progressList)
	newStatuses := computeNodeStatuses(courses, prereqs, pathEdges, statusMap)

	newlyAvailable := []dto.StatusChangeItem{}
	for _, c := range courses {
		if c.ID == courseID {
			continue
		}
		if nodeStatuses[c.ID] == entity.NodeStatusLocked && newStatuses[c.ID] == entity.NodeStatusAvailable {
			newlyAvailable = append(newlyAvailable, dto.StatusChangeItem{
				CourseID:   c.ID.String(),
				CourseCode: c.Code,
				CourseName: c.Name,
				Status:     "AVAILABLE",
			})
		}
	}

	return dto.ClaimCourseResponse{NewlyAvailable: newlyAvailable}, nil
}

// =========== UNCLAIM ===========

func (s *skillTreeService) UnclaimCourse(ctx context.Context, userID, courseID uuid.UUID) (dto.UnclaimCourseResponse, error) {
	existing, err := s.skillTreeRepo.GetProgressByCourseAndUser(ctx, userID, courseID)
	if err != nil || existing == nil {
		return dto.UnclaimCourseResponse{}, dto.ErrProgressNotFound
	}
	if existing.Status != entity.NodeStatusCompleted {
		return dto.UnclaimCourseResponse{}, dto.ErrCourseNotCompleted
	}

	courses, _ := s.skillTreeRepo.GetAllCourses(ctx)
	prereqs, _ := s.skillTreeRepo.GetAllPrerequisites(ctx)
	pathEdges, _ := s.skillTreeRepo.GetAllPathEdges(ctx)
	progressList, _ := s.skillTreeRepo.GetProgressByUserID(ctx, userID)
	statusMap, _, _ := buildStatusMaps(progressList)
	oldStatuses := computeNodeStatuses(courses, prereqs, pathEdges, statusMap)

	if err := s.skillTreeRepo.DeleteProgress(ctx, userID, courseID); err != nil {
		return dto.UnclaimCourseResponse{}, err
	}

	progressList, _ = s.skillTreeRepo.GetProgressByUserID(ctx, userID)
	statusMap, _, _ = buildStatusMaps(progressList)
	newStatuses := computeNodeStatuses(courses, prereqs, pathEdges, statusMap)

	newlyLocked := []dto.StatusChangeItem{}
	for _, c := range courses {
		if c.ID == courseID {
			continue
		}
		if oldStatuses[c.ID] == entity.NodeStatusAvailable && newStatuses[c.ID] == entity.NodeStatusLocked {
			newlyLocked = append(newlyLocked, dto.StatusChangeItem{
				CourseID:   c.ID.String(),
				CourseCode: c.Code,
				CourseName: c.Name,
				Status:     "LOCKED",
			})
		}
	}

	return dto.UnclaimCourseResponse{NewlyLocked: newlyLocked}, nil
}

// =========== HELPERS ===========

func buildStatusMaps(progressList []entity.StudentProgress) (
	map[uuid.UUID]entity.NodeStatus,
	map[uuid.UUID]*string,
	map[uuid.UUID]*time.Time,
) {
	statusMap := map[uuid.UUID]entity.NodeStatus{}
	gradeMap := map[uuid.UUID]*string{}
	claimedAtMap := map[uuid.UUID]*time.Time{}
	for _, p := range progressList {
		statusMap[p.CourseID] = p.Status
		gradeMap[p.CourseID] = p.Grade
		claimedAtMap[p.CourseID] = p.ClaimedAt
	}
	return statusMap, gradeMap, claimedAtMap
}

// computeNodeStatuses determines COMPLETED/AVAILABLE/LOCKED for each course.
// A course is AVAILABLE if all its prerequisites (via prerequisites table) are COMPLETED
// AND all its path_edge incoming nodes are COMPLETED.
func computeNodeStatuses(
	courses []entity.Course,
	prereqs []entity.Prerequisite,
	pathEdges []entity.PathEdge,
	completedMap map[uuid.UUID]entity.NodeStatus,
) map[uuid.UUID]entity.NodeStatus {
	// Build: courseID -> list of required courseIDs (prerequisites)
	prereqMap := map[uuid.UUID][]uuid.UUID{}
	for _, p := range prereqs {
		prereqMap[p.CourseID] = append(prereqMap[p.CourseID], p.RequireID)
	}

	// Build: courseID -> list of required path source courseIDs
	pathInMap := map[uuid.UUID][]uuid.UUID{}
	for _, e := range pathEdges {
		pathInMap[e.ToCourseID] = append(pathInMap[e.ToCourseID], e.FromCourseID)
	}

	result := map[uuid.UUID]entity.NodeStatus{}

	for _, c := range courses {
		if completedMap[c.ID] == entity.NodeStatusCompleted {
			result[c.ID] = entity.NodeStatusCompleted
			continue
		}

		allPrereqMet := true
		for _, reqID := range prereqMap[c.ID] {
			if completedMap[reqID] != entity.NodeStatusCompleted {
				allPrereqMet = false
				break
			}
		}

		allPathMet := true
		for _, srcID := range pathInMap[c.ID] {
			if completedMap[srcID] != entity.NodeStatusCompleted {
				allPathMet = false
				break
			}
		}

		if allPrereqMet && allPathMet {
			result[c.ID] = entity.NodeStatusAvailable
		} else {
			result[c.ID] = entity.NodeStatusLocked
		}
	}

	return result
}

// bfsIDs performs BFS from startID using the given adjacency map, returns all visited IDs (excluding start).
func bfsIDs(startID uuid.UUID, adjMap map[uuid.UUID][]uuid.UUID) map[uuid.UUID]bool {
	visited := map[uuid.UUID]bool{startID: true}
	queue := []uuid.UUID{startID}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, neighbor := range adjMap[curr] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	delete(visited, startID)
	return visited
}

func courseToGraphNode(c entity.Course) dto.GraphNodeResponse {
	labPaths := []dto.LabPathResponse{}
	if c.LabPath != nil {
		labPaths = append(labPaths, dto.LabPathResponse{
			ID:    c.LabPath.ID.String(),
			Name:  c.LabPath.Name,
			Color: c.LabPath.Color,
		})
	}
	return dto.GraphNodeResponse{
		ID:          c.ID.String(),
		Code:        c.Code,
		Name:        c.Name,
		Credit:      c.Credit,
		Semester:    c.Semester,
		IsElective:  c.IsElective,
		Description: c.Description,
		LabPaths:    labPaths,
	}
}

func buildEdges(prereqs []entity.Prerequisite, pathEdges []entity.PathEdge) []dto.GraphEdgeResponse {
	edges := make([]dto.GraphEdgeResponse, 0, len(prereqs)+len(pathEdges))

	for _, p := range prereqs {
		edges = append(edges, dto.GraphEdgeResponse{
			ID:     p.ID.String(),
			Source: p.RequireID.String(),
			Target: p.CourseID.String(),
			Type:   "PREREQUISITE",
			Color:  "#2C2C2A",
		})
	}
	for _, e := range pathEdges {
		edges = append(edges, dto.GraphEdgeResponse{
			ID:     e.ID.String(),
			Source: e.FromCourseID.String(),
			Target: e.ToCourseID.String(),
			Type:   "PATH",
			Color:  "#1D9E75",
		})
	}
	return edges
}
