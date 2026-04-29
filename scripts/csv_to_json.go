package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Course struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Credit      int     `json:"credit"`
	Semester    int     `json:"semester"`
	IsElective  bool    `json:"is_elective"`
	Description *string `json:"description"`
	Lab         string  `json:"lab"`
}

func main() {
	inputPath := "course.csv"
	outputPath := "database/json/courses.json"

	if len(os.Args) > 1 {
		inputPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		outputPath = os.Args[2]
	}
	if len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/csv_to_json.go [input.csv] [output.json]")
		os.Exit(2)
	}

	if err := convertCSVToJSON(inputPath, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote %s\n", outputPath)
}

func convertCSVToJSON(inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	r := csv.NewReader(file)
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if err == io.EOF {
		return fmt.Errorf("empty CSV: %s", inputPath)
	}
	if err != nil {
		return err
	}

	// Normalize header names so column order does not matter.
	header[0] = strings.TrimPrefix(header[0], "\ufeff")
	index := map[string]int{}
	for i, name := range header {
		key := strings.ToLower(strings.TrimSpace(name))
		if key != "" {
			index[key] = i
		}
	}

	required := []string{"id", "code", "name", "credit", "semester", "is_elective", "lab"}
	for _, key := range required {
		if _, ok := index[key]; !ok {
			return fmt.Errorf("missing column %q in header", key)
		}
	}

	items := make([]Course, 0, 64)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if isEmptyRecord(record) {
			continue
		}

		get := func(key string) string {
			idx, ok := index[key]
			if !ok || idx >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[idx])
		}

		credit, err := parseInt(get("credit"))
		if err != nil {
			return fmt.Errorf("invalid credit %q: %w", get("credit"), err)
		}

		semester, err := parseInt(get("semester"))
		if err != nil {
			return fmt.Errorf("invalid semester %q: %w", get("semester"), err)
		}

		isElective, err := parseBool(get("is_elective"))
		if err != nil {
			return fmt.Errorf("invalid is_elective %q: %w", get("is_elective"), err)
		}

		description := strings.TrimSpace(get("description"))
		var descriptionPtr *string
		if description != "" {
			descriptionPtr = &description
		}

		item := Course{
			ID:          get("id"),
			Code:        get("code"),
			Name:        get("name"),
			Credit:      credit,
			Semester:    semester,
			IsElective:  isElective,
			Description: descriptionPtr,
			Lab:         get("lab"),
		}

		if item.ID == "" || item.Code == "" || item.Name == "" || item.Lab == "" {
			return fmt.Errorf("empty required field in record: %v", record)
		}

		items = append(items, item)
	}

	payload, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	payload = append(payload, '\n')
	return os.WriteFile(outputPath, payload, 0o644)
}

func isEmptyRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func parseInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	return strconv.Atoi(value)
}

func parseBool(value string) (bool, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return false, fmt.Errorf("empty value")
	}

	switch value {
	case "true", "t", "1", "yes", "y":
		return true, nil
	case "false", "f", "0", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("unknown boolean")
	}
}
