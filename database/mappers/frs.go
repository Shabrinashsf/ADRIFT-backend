package mappers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type JadwalKuliah struct {
	Hari       string `json:"hari"`
	Jam        string `json:"jam"`
	Ruangan    string `json:"ruangan"`
	Prodi      string `json:"prodi"`
	MataKuliah string `json:"mata_kuliah"`
	Semester   string `json:"semester"`
	KodeDosen  string `json:"kode_dosen"`
	SKS        string `json:"sks"`
	RawData    string `json:"raw_data"`
}

type ScheduleOutput struct {
	ID           uuid.UUID           `json:"id"`
	CourseName   string              `json:"course_name"`
	LectureCode  string              `json:"lecture_code"`
	Class        string              `json:"class"`
	Day          entity.Day          `json:"day"`
	StartTime    time.Time           `json:"start_time"`
	EndTime      time.Time           `json:"end_time"`
	Room         string              `json:"room"`
	Semester     int                 `json:"semester"`
	AcademicYear string              `json:"academic_year"`
	Capacity     int                 `json:"capacity"`
	SKS          int                 `json:"sks"`
	Prodi        entity.ProdiType    `json:"prodi"`
	Term         entity.TermSemester `json:"term"`
}

type ScheduleNullSummary struct {
	CourseNames []string `json:"course_names"`
}

type ScheduleNullReport struct {
	Records []map[string]any    `json:"records"`
	Summary ScheduleNullSummary `json:"summary"`
}

var scheduleFields = []string{
	"id",
	"course_name",
	"lecture_code",
	"class",
	"day",
	"start_time",
	"end_time",
	"room",
	"semester",
	"academic_year",
	"capacity",
	"sks",
	"prodi",
	"term",
}

const scheduleSheetName = "Jadwal Kuliah"

func GenerateScheduleFiles(filePath, outputJSON, reportJSON, academicYear string, term entity.TermSemester) (ScheduleNullReport, error) {
	jadwalByProdiSemester, _, _, err := mapExcelToSchedule(filePath, scheduleSheetName, term, academicYear)
	if err != nil {
		return ScheduleNullReport{}, err
	}

	report, err := writeScheduleFiles(jadwalByProdiSemester, outputJSON, reportJSON)
	if err != nil {
		return ScheduleNullReport{}, err
	}

	return report, nil
}

func mapExcelToSchedule(filePath, sheetName string, term entity.TermSemester, academicYear string) (map[string]map[int][]ScheduleOutput, map[string]map[int]int, []ScheduleOutput, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gagal membuka file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gagal membaca rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, nil, nil, fmt.Errorf("file excel kosong atau tidak ada data")
	}

	// baris 1 (index 0): Header ruangan (IF-101, IF-102, dll)
	headerRuangan := rows[0]

	// mapping semua jadwal
	var semuaJadwal []ScheduleOutput
	jadwalCountByProdiSemester := make(map[string]map[int]int)
	jadwalByProdiSemester := make(map[string]map[int][]ScheduleOutput)

	// mapping hari
	hariList := []string{"SENIN", "SELASA", "RABU", "KAMIS", "JUMAT"}
	currentHari := ""

	// mulai dari baris ke-2 (index 1)
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if len(row) == 0 {
			continue
		}

		// Kolom A: Hari (jika ada, update currentHari)
		if len(row) > 0 {
			cellA := strings.ToUpper(strings.TrimSpace(row[0]))
			for _, h := range hariList {
				if cellA == h {
					currentHari = h
					break
				}
			}
		}

		// Kolom B: Jam
		jamBaris1 := ""
		if len(row) > 1 {
			jamBaris1 = strings.TrimSpace(row[1])
		}

		if jamBaris1 == "" || currentHari == "" {
			continue
		}

		// Kolom C-F (index 2-5): Parse SKPB Reguler dan SKPB IUP
		for colIdx := 2; colIdx <= 5; colIdx++ {
			cellValue := ""
			if colIdx < len(row) {
				cellValue = strings.TrimSpace(row[colIdx])
			}
			if cellValue == "" {
				continue
			}

			// Ambil info SKS/SEM dari baris berikutnya
			infoCell := ""
			if rowIdx+1 < len(rows) {
				nextRow := rows[rowIdx+1]
				if colIdx < len(nextRow) {
					infoCell = strings.TrimSpace(nextRow[colIdx])
				}
			}

			// Parse SKPB dan dapatkan list jadwal (bisa multiple prodi)
			rawData := cellValue + "\n" + infoCell
			jadwalSKPBList := parseSKPB(currentHari, jamBaris1, "", cellValue, infoCell, rawData)

			for _, jadwal := range jadwalSKPBList {
				schedule, ok := toScheduleOutput(jadwal, term, academicYear)
				if !ok {
					continue
				}

				semuaJadwal = append(semuaJadwal, schedule)
				addScheduleCount(jadwalCountByProdiSemester, string(schedule.Prodi), schedule.Semester)
				addScheduleGroup(jadwalByProdiSemester, schedule)
			}
		}

		// Kolom G ke atas (index 6+): Data mata kuliah per ruangan
		for colIdx := 6; colIdx < len(row); colIdx++ {
			cellValue := strings.TrimSpace(row[colIdx])
			if cellValue == "" {
				continue
			}

			// Ini baris nama matkul, ambil baris berikutnya untuk info sem/dosen/sks
			semInfo := ""
			jamBaris2 := ""
			if rowIdx+1 < len(rows) {
				nextRow := rows[rowIdx+1]
				if colIdx < len(nextRow) {
					semInfo = strings.TrimSpace(nextRow[colIdx])
				}
				// Ambil jam dari baris berikutnya
				if len(nextRow) > 1 {
					jamBaris2 = strings.TrimSpace(nextRow[1])
				}
			}

			// Extract jam mulai dan jam selesai, lalu gabungkan
			jamMulai := extractJamMulai(jamBaris1)
			jamSelesai := extractJamSelesai(jamBaris2)
			jam := strings.TrimSpace(jamMulai)
			if jamSelesai != "" {
				jam = fmt.Sprintf("%s - %s", jamMulai, jamSelesai)
			}

			// Ambil nama ruangan dari header dan bersihkan
			ruangan := ""
			if colIdx < len(headerRuangan) {
				ruangan = cleanRuangan(strings.TrimSpace(headerRuangan[colIdx]))
			}

			// Parse data mata kuliah
			rawData := cellValue + "\n" + semInfo
			jadwal := parseJadwal(currentHari, jam, ruangan, cellValue, semInfo, rawData)

			schedule, ok := toScheduleOutput(jadwal, term, academicYear)
			if !ok {
				continue
			}

			semuaJadwal = append(semuaJadwal, schedule)
			addScheduleCount(jadwalCountByProdiSemester, string(schedule.Prodi), schedule.Semester)
			addScheduleGroup(jadwalByProdiSemester, schedule)
		}
	}

	return jadwalByProdiSemester, jadwalCountByProdiSemester, semuaJadwal, nil
}

func writeScheduleFiles(jadwalByProdiSemester map[string]map[int][]ScheduleOutput, outputJSON, reportJSON string) (ScheduleNullReport, error) {
	jsonData, err := json.MarshalIndent(jadwalByProdiSemester, "", "  ")
	if err != nil {
		return ScheduleNullReport{}, fmt.Errorf("gagal convert ke JSON: %w", err)
	}

	if err := os.WriteFile(outputJSON, jsonData, 0644); err != nil {
		return ScheduleNullReport{}, fmt.Errorf("gagal menulis file: %w", err)
	}

	report, err := generateScheduleNullReport(outputJSON, reportJSON)
	if err != nil {
		return ScheduleNullReport{}, err
	}

	return report, nil
}

func MapperFRS() {
	filePath := "database/mappers/jadwal_frs_genap_2025_2026.xlsx"
	term := extractTermFromFilePath(filePath)
	academicYear := extractAcademicYearFromFilePath(filePath)

	sheetName := scheduleSheetName
	fmt.Printf("Membaca sheet: %s\n\n", sheetName)

	jadwalByProdiSemester, jadwalCountByProdiSemester, semuaJadwal, err := mapExcelToSchedule(filePath, sheetName, term, academicYear)
	if err != nil {
		log.Fatalf("Gagal memproses file: %v", err)
	}

	// output hasil
	fmt.Println("=== SEMUA JADWAL ===")
	fmt.Printf("Total: %d jadwal ditemukan\n\n", len(semuaJadwal))

	fmt.Println("=== JADWAL PER PRODI & SEMESTER ===")
	for prodi, semesters := range jadwalCountByProdiSemester {
		fmt.Printf("  %s:\n", prodi)
		for sem, count := range semesters {
			fmt.Printf("    Semester %d: %d mata kuliah\n", sem, count)
		}
	}

	outputFileJSON := "database/json/schedule.json"
	reportFileJSON := "database/json/schedule_null_report.json"

	if _, err := writeScheduleFiles(jadwalByProdiSemester, outputFileJSON, reportFileJSON); err != nil {
		log.Fatalf("Gagal membuat file output: %v", err)
	}

	fmt.Printf("\n✓ Berhasil export ke %s\n", outputFileJSON)
	fmt.Printf("✓ Berhasil export ke %s\n", reportFileJSON)
}

// cleanRuangan membersihkan nama ruangan dari suffix seperti "a&b"
// "IF-105a&b (kapasitas 100)" -> "IF-105 (kapasitas 100)"
func cleanRuangan(ruangan string) string {
	// Regex untuk menghapus suffix seperti "a&b", "a&b&c", dll sebelum spasi atau kurung
	re := regexp.MustCompile(`([A-Z]+-\d+)[a-z&]+\s*(\(.*\))?`)
	if re.MatchString(ruangan) {
		return re.ReplaceAllString(ruangan, "$1 $2")
	}
	return ruangan
}

func toScheduleOutput(jadwal JadwalKuliah, term entity.TermSemester, academicYear string) (ScheduleOutput, bool) {
	if strings.TrimSpace(jadwal.MataKuliah) == "" {
		return ScheduleOutput{}, false
	}

	startTime, endTime, ok := parseTimeRange(jadwal.Jam)
	if !ok {
		return ScheduleOutput{}, false
	}

	semester := parseInt(jadwal.Semester)
	sks := parseInt(jadwal.SKS)

	schedule := ScheduleOutput{
		ID:           uuid.New(),
		CourseName:   normalizeCourseName(jadwal.MataKuliah),
		LectureCode:  strings.TrimSpace(jadwal.KodeDosen),
		Class:        extractClass(jadwal.MataKuliah),
		Day:          normalizeDay(jadwal.Hari),
		StartTime:    startTime,
		EndTime:      endTime,
		Room:         stripCapacityFromRoom(jadwal.Ruangan),
		Semester:     semester,
		AcademicYear: academicYear,
		Capacity:     extractCapacity(jadwal.Ruangan),
		SKS:          sks,
		Prodi:        entity.ProdiType(strings.TrimSpace(jadwal.Prodi)),
		Term:         term,
	}

	return schedule, true
}

func addScheduleCount(counter map[string]map[int]int, prodi string, semester int) {
	if strings.TrimSpace(prodi) == "" || semester == 0 {
		return
	}

	if _, ok := counter[prodi]; !ok {
		counter[prodi] = make(map[int]int)
	}

	counter[prodi][semester]++
}

func addScheduleGroup(target map[string]map[int][]ScheduleOutput, schedule ScheduleOutput) {
	prodi := strings.TrimSpace(string(schedule.Prodi))
	if prodi == "" || schedule.Semester == 0 {
		return
	}

	if _, ok := target[prodi]; !ok {
		target[prodi] = make(map[int][]ScheduleOutput)
	}

	target[prodi][schedule.Semester] = append(target[prodi][schedule.Semester], schedule)
}

func extractTermFromFilePath(filePath string) entity.TermSemester {
	lower := strings.ToLower(filePath)
	if strings.Contains(lower, "genap") {
		return entity.TermSemesterGenap
	}
	if strings.Contains(lower, "ganjil") {
		return entity.TermSemesterGanjil
	}
	return ""
}

func extractAcademicYearFromFilePath(filePath string) string {
	baseName := filepath.Base(filePath)
	re := regexp.MustCompile(`(\d{4})[^\d]?(\d{4})`)
	match := re.FindStringSubmatch(baseName)
	if len(match) == 3 {
		return fmt.Sprintf("%s/%s", match[1], match[2])
	}
	return ""
}

func normalizeDay(day string) entity.Day {
	switch strings.ToUpper(strings.TrimSpace(day)) {
	case "SENIN":
		return entity.DaySenin
	case "SELASA":
		return entity.DaySelasa
	case "RABU":
		return entity.DayRabu
	case "KAMIS":
		return entity.DayKamis
	case "JUMAT":
		return entity.DayJumat
	case "SABTU":
		return entity.DaySabtu
	case "MINGGU":
		return entity.DayMinggu
	default:
		return entity.Day(strings.TrimSpace(day))
	}
}

func parseTimeRange(jam string) (time.Time, time.Time, bool) {
	parts := strings.Split(jam, "-")
	if len(parts) == 0 {
		return time.Time{}, time.Time{}, false
	}

	startTime, ok := parseClock(strings.TrimSpace(parts[0]))
	if !ok {
		return time.Time{}, time.Time{}, false
	}

	if len(parts) == 1 {
		return startTime, startTime, true
	}

	endTime, ok := parseClock(strings.TrimSpace(parts[1]))
	if !ok {
		return time.Time{}, time.Time{}, false
	}

	return startTime, endTime, true
}

func parseClock(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}

	normalized := strings.ReplaceAll(strings.TrimSpace(value), ":", ".")
	parsed, err := time.Parse("15.04", normalized)
	if err != nil {
		return time.Time{}, false
	}

	return time.Date(2000, 1, 1, parsed.Hour(), parsed.Minute(), 0, 0, time.Local), true
}

func extractClass(mataKuliah string) string {
	cleaned := strings.TrimSpace(mataKuliah)
	cleaned = regexp.MustCompile(`\s*\(EN\)\s*$`).ReplaceAllString(cleaned, "")
	parts := strings.Fields(cleaned)
	if len(parts) == 0 {
		return ""
	}

	last := parts[len(parts)-1]
	if regexp.MustCompile(`^[A-Z]$`).MatchString(last) {
		return last
	}

	return ""
}

func extractCapacity(ruangan string) int {
	re := regexp.MustCompile(`(?i)kapasitas\s*(\d+)`)
	match := re.FindStringSubmatch(ruangan)
	if len(match) < 2 {
		return 0
	}

	capacity, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}

	return capacity
}

func stripCapacityFromRoom(ruangan string) string {
	re := regexp.MustCompile(`(?i)\s*\(\s*kapasitas\s*\d+\s*\)`)
	cleaned := re.ReplaceAllString(ruangan, "")
	return strings.TrimSpace(strings.Join(strings.Fields(cleaned), " "))
}

func normalizeCourseName(mataKuliah string) string {
	cleaned := strings.TrimSpace(mataKuliah)
	cleaned = regexp.MustCompile(`\s*\(EN\)\s*$`).ReplaceAllString(cleaned, "")
	parts := strings.Fields(cleaned)
	if len(parts) == 0 {
		return ""
	}

	last := parts[len(parts)-1]
	if regexp.MustCompile(`^[A-Z]$`).MatchString(last) {
		parts = parts[:len(parts)-1]
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

func generateScheduleNullReport(inputPath, outputPath string) (ScheduleNullReport, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return ScheduleNullReport{}, fmt.Errorf("gagal membaca file: %w", err)
	}

	var raw map[string]map[string][]map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return ScheduleNullReport{}, fmt.Errorf("gagal parse JSON: %w", err)
	}

	report := ScheduleNullReport{}
	seenCourseNames := make(map[string]struct{})

	for prodi, semesterMap := range raw {
		for semester, items := range semesterMap {
			for _, item := range items {
				if item == nil {
					report.Records = append(report.Records, map[string]any{
						"prodi":    prodi,
						"semester": semester,
					})
					continue
				}

				if recordHasNullLike(item) {
					report.Records = append(report.Records, item)
					addCourseNameSummary(&report, seenCourseNames, item)
				}
			}
		}
	}

	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return ScheduleNullReport{}, fmt.Errorf("gagal membuat report JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return ScheduleNullReport{}, fmt.Errorf("gagal menulis report: %w", err)
	}

	return report, nil
}

func recordHasNullLike(item map[string]any) bool {
	for _, field := range scheduleFields {
		value, ok := item[field]
		if !ok {
			return true
		}
		if value == nil {
			return true
		}
		if strValue, ok := value.(string); ok {
			if strings.TrimSpace(strValue) == "" {
				return true
			}
		}
	}
	return false
}

func addCourseNameSummary(report *ScheduleNullReport, seen map[string]struct{}, record map[string]any) {
	value, ok := record["course_name"]
	if !ok {
		return
	}

	name, ok := value.(string)
	if !ok {
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	if _, exists := seen[name]; exists {
		return
	}

	seen[name] = struct{}{}
	report.Summary.CourseNames = append(report.Summary.CourseNames, name)
}

func parseInt(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return parsed
}

// parseJadwal mengekstrak informasi dari cell jadwal
func parseJadwal(hari, jam, ruangan, matkulCell, semInfoCell, rawData string) JadwalKuliah {
	jadwal := JadwalKuliah{
		Hari:    hari,
		Jam:     jam,
		Ruangan: ruangan,
		RawData: rawData,
	}

	// Parse baris 1: PRODI_NamaMataKuliah
	prodiRegex := regexp.MustCompile(`^(IF|IUP|RKA|RPL|S3)_(.+)$`)
	prodiMatch := prodiRegex.FindStringSubmatch(matkulCell)
	if len(prodiMatch) == 3 {
		jadwal.Prodi = prodiMatch[1]
		jadwal.MataKuliah = strings.TrimSpace(prodiMatch[2])
	} else {
		// Jika tidak sesuai format standar, anggap sebagai mata kuliah S2
		jadwal.Prodi = "S2"
		jadwal.MataKuliah = strings.TrimSpace(matkulCell)
	}

	// Parse baris 2: Sem X / KODE_DOSEN / Y SKS
	// Format: "Sem 7 / SL / 3 SKS" atau "Sem 4 / SR, MA / 3 SKS"
	if semInfoCell != "" {
		parts := strings.Split(semInfoCell, "/")
		if len(parts) >= 3 {
			// Part 1: Semester
			semPart := strings.TrimSpace(parts[0])
			semRegex := regexp.MustCompile(`(?i)sem\s*(\d+)`)
			semMatch := semRegex.FindStringSubmatch(semPart)
			if len(semMatch) > 1 {
				jadwal.Semester = semMatch[1]
			}

			// Part 2: Kode Dosen (2 huruf, bisa multiple dipisah koma)
			jadwal.KodeDosen = strings.TrimSpace(parts[1])

			// Part 3: SKS
			sksPart := strings.TrimSpace(parts[2])
			sksRegex := regexp.MustCompile(`(\d+)\s*SKS`)
			sksMatch := sksRegex.FindStringSubmatch(sksPart)
			if len(sksMatch) > 1 {
				jadwal.SKS = sksMatch[1]
			}
		}
	}

	return jadwal
}

// extractJamMulai mengambil jam mulai dari format "07.00 - 07.50"
func extractJamMulai(jam string) string {
	jam = strings.ReplaceAll(jam, " ", "")
	if parts := strings.Split(jam, "-"); len(parts) >= 1 {
		return strings.TrimSpace(parts[0])
	}
	return jam
}

// extractJamSelesai mengambil jam selesai dari format "08.00 - 08.50"
func extractJamSelesai(jam string) string {
	jam = strings.ReplaceAll(jam, " ", "")
	if parts := strings.Split(jam, "-"); len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return jam
}

// parseSKPB mengekstrak informasi dari cell SKPB
// Format mata kuliah: "KALKULUS 2 - IF & RKA" (bisa multiple prodi dipisah &)
// Format info: "3 SKS / SEM 2"
// Mengembalikan slice JadwalKuliah karena bisa ada multiple prodi
func parseSKPB(hari, jam, ruangan, skpbCell, infoCell, rawData string) []JadwalKuliah {
	var jadwalList []JadwalKuliah

	if strings.TrimSpace(skpbCell) == "" {
		return jadwalList
	}

	// Parse baris 1: MataKuliah - Prodi1 & Prodi2 ...
	parts := strings.Split(skpbCell, "-")
	if len(parts) < 2 {
		return jadwalList
	}

	mataKuliah := strings.TrimSpace(parts[0])
	prodiPart := strings.TrimSpace(parts[1])

	// Split prodi yang dipisahkan dengan &
	prodis := strings.Split(prodiPart, "&")

	// Parse baris 2: SKS / SEM
	var sks, semester string
	if infoCell != "" {
		infoParts := strings.Split(infoCell, "/")
		if len(infoParts) >= 2 {
			// Part 1: SKS (format: "3 SKS")
			sksPart := strings.TrimSpace(infoParts[0])
			sksRegex := regexp.MustCompile(`(\d+)\s*SKS`)
			sksMatch := sksRegex.FindStringSubmatch(sksPart)
			if len(sksMatch) > 1 {
				sks = sksMatch[1]
			}

			// Part 2: SEM (format: "SEM 2")
			semPart := strings.TrimSpace(infoParts[1])
			semRegex := regexp.MustCompile(`(?i)sem\s*(\d+)`)
			semMatch := semRegex.FindStringSubmatch(semPart)
			if len(semMatch) > 1 {
				semester = semMatch[1]
			}
		}
	}

	// Buat jadwal untuk setiap prodi
	for _, prodi := range prodis {
		prodi = strings.TrimSpace(prodi)
		if prodi == "" {
			continue
		}

		jadwal := JadwalKuliah{
			Hari:       hari,
			Jam:        jam,
			Ruangan:    ruangan,
			Prodi:      prodi,
			MataKuliah: mataKuliah,
			Semester:   semester,
			SKS:        sks,
			RawData:    rawData,
		}

		jadwalList = append(jadwalList, jadwal)
	}

	return jadwalList
}
