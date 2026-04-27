package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProdiType string

type Day string

type TermSemester string

const (
	ProdiTypeS1Informatika    ProdiType = "IF"
	ProdiTypeS1InformatikaIUP ProdiType = "IUP"
	ProdiTypeS1RKA            ProdiType = "RKA"
	ProdiTypeS1RPL            ProdiType = "RPL"
	ProdiTypeS2Informatika    ProdiType = "S2"
	ProdiTypeS3Informatika    ProdiType = "S3"
)

const (
	DaySenin  Day = "Senin"
	DaySelasa Day = "Selasa"
	DayRabu   Day = "Rabu"
	DayKamis  Day = "Kamis"
	DayJumat  Day = "Jumat"
	DaySabtu  Day = "Sabtu"
	DayMinggu Day = "Minggu"
)

const (
	TermSemesterGanjil TermSemester = "GANJIL"
	TermSemesterGenap  TermSemester = "GENAP"
)

type Schedule struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;" json:"id"`
	CourseID     uuid.UUID    `json:"course_id"`
	LectureID    uuid.UUID    `json:"lecture_id"`
	Class        string       `json:"class"` // A, B, C, D, E, dll
	Day          Day          `json:"day"`
	StartTime    time.Time    `json:"start_time"`
	EndTime      time.Time    `json:"end_time"`
	Room         string       `json:"room"`
	Semester     int          `json:"semester"`
	AcademicYear string       `json:"academic_year"`
	Capacity     int          `json:"capacity"`
	SKS          int          `json:"sks"`
	Prodi        ProdiType    `json:"prodi"`
	Term         TermSemester `json:"term"`

	FRSPlanItems []FRSPlanItem `json:"frs_plan_items,omitempty"`

	Lecture *Lecture `gorm:"foreignKey:LectureID"`
	Course  *Course  `gorm:"foreignKey:CourseID"`
}

func (s *Schedule) TableName() string {
	return "schedules"
}
