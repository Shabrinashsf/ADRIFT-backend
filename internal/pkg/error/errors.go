package myerror

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var (
	ErrGeneral     = New("something went wrong", 500)
	ErrBodyRequest = New("failed get body request", 400)
)

type Error struct {
	Message    string
	StatusCode int
}

func New(msg string, statusCode int) Error {
	return Error{
		Message:    msg,
		StatusCode: statusCode,
	}
}

func (e Error) Error() string {
	return e.Message
}

// Convert validation error to user-friendly message
func FormatValidationError(err error) string {
	if err == nil {
		return "invalid request"
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		if len(ve) > 0 {
			firstErr := ve[0]
			fieldName := toSnakeCase(firstErr.Field())

			switch firstErr.Tag() {
			case "required":
				return fieldName + " is required"
			case "email":
				return fieldName + " must be a valid email"
			case "min":
				return fieldName + " is too short (minimum: " + firstErr.Param() + ")"
			case "max":
				return fieldName + " is too long (maximum: " + firstErr.Param() + ")"
			case "numeric":
				return fieldName + " must be numeric"
			case "gte":
				return fieldName + " must be at least " + firstErr.Param()
			case "lte":
				return fieldName + " must be at most " + firstErr.Param()
			default:
				return fieldName + " failed validation: " + firstErr.Tag()
			}
		}
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "json:") {
		return "invalid JSON format in request body"
	}

	return "invalid request data"
}

func toSnakeCase(s string) string {
	snake := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(snake, "${1}_${2}"))
}

// Convert DB error (GORM, PostgreSQL) to user-friendly message
func FromDBError(err error) Error {
	if err == nil {
		return ErrGeneral
	}

	// Check for specific GORM errors
	if err == gorm.ErrRecordNotFound {
		return New("record not found", 404)
	}

	userMsg := parseDBError(err)

	// Determine appropriate status code based on the parsed error
	statusCode := determineStatusCodeFromError(err)

	return New(userMsg, statusCode)
}

func parseDBError(err error) string {
	if err == nil {
		return "database error occurred"
	}

	errMsg := err.Error()

	// PostgreSQL error patterns
	// Format: ERROR: <error message> (SQLSTATE <code>)
	pgErrorPattern := regexp.MustCompile(`ERROR:\s*(.+?)\s*\(SQLSTATE\s+(\d+)\)`)
	matches := pgErrorPattern.FindStringSubmatch(errMsg)

	if len(matches) < 3 {
		// Not a PostgreSQL error, return original message
		return errMsg
	}

	errorMessage := matches[1]
	sqlState := matches[2]

	// Handle common SQLSTATE codes
	switch sqlState {
	case "23505": // unique_violation
		return parseUniqueViolationError(errorMessage)
	case "23503": // foreign_key_violation
		return parseForeignKeyViolationError(errorMessage)
	case "23502": // not_null_violation
		return parseNotNullViolationError(errorMessage)
	case "22P02": // invalid_text_representation
		return "invalid data format provided"
	case "08001": // connection_does_not_exist
		return "database connection error"
	case "28000": // invalid_authorization_specification
		return "database authentication failed"
	default:
		// Return a generic message for unknown SQLSTATE codes
		return errorMessage
	}
}

// Parses unique constraint violation errors
func parseUniqueViolationError(errorMessage string) string {
	// Extract constraint name from error message
	// Pattern: duplicate key value violates unique constraint "constraint_name"
	constraintPattern := regexp.MustCompile(`unique constraint\s+"?(\w+)"?`)
	matches := constraintPattern.FindStringSubmatch(errorMessage)

	if len(matches) < 2 {
		return "a record with this value already exists"
	}

	constraintName := matches[1]

	// Parse constraint name to extract field name
	// Common patterns:
	// - idx_users_phone_number → phone_number
	// - users_email_key → email
	// - unique_users_email → email
	fieldName := extractFieldNameFromConstraint(constraintName)

	if fieldName != "" {
		return fieldName + " already exists"
	}

	return "a record with this value already exists"
}

// Parses foreign key constraint violation errors
// Example: "insert or update on table \"users\" violates foreign key constraint \"fk_users_class_id_classes\""
// Returns: "referenced resource does not exist"
func parseForeignKeyViolationError(errorMessage string) string {
	fkPattern := regexp.MustCompile(`on table\s+"?(\w+)"?\s+violates foreign key constraint\s+"?(\w+)"?`)
	matches := fkPattern.FindStringSubmatch(errorMessage)

	if len(matches) >= 2 {
		tableName := matches[1]
		return "invalid reference to " + tableName
	}

	return "referenced resource does not exist"
}

// Parses not null constraint violation errors
// Example: "null value in column \"email\" violates not-null constraint"
// Returns: "email is required"
func parseNotNullViolationError(errorMessage string) string {
	columnPattern := regexp.MustCompile(`null value in column\s+"?(\w+)"?`)
	matches := columnPattern.FindStringSubmatch(errorMessage)

	if len(matches) >= 2 {
		columnName := matches[1]
		return columnName + " is required"
	}

	return "required field is missing"
}

// Extracts field name from various constraint naming patterns
func extractFieldNameFromConstraint(constraintName string) string {
	// Pattern 1: idx_tablename_fieldname (e.g., idx_users_phone_number)
	// Use [a-z]+ for table name to avoid matching underscores in field names
	pattern1 := regexp.MustCompile(`^idx_[a-z]+_(.+)$`)
	if matches := pattern1.FindStringSubmatch(constraintName); len(matches) >= 2 {
		return matches[1]
	}

	// Pattern 2: tablename_fieldname_key (e.g., users_email_key)
	// Use [a-z]+ for table name
	pattern2 := regexp.MustCompile(`^[a-z]+_(.+)_key$`)
	if matches := pattern2.FindStringSubmatch(constraintName); len(matches) >= 2 {
		return matches[1]
	}

	// Pattern 3: unique_tablename_fieldname (e.g., unique_users_email)
	pattern3 := regexp.MustCompile(`^unique_[a-z]+_(.+)$`)
	if matches := pattern3.FindStringSubmatch(constraintName); len(matches) >= 2 {
		return matches[1]
	}

	// Pattern 4: tablename_fieldname_unique (e.g., users_email_unique)
	pattern4 := regexp.MustCompile(`^[a-z]+_(.+)_unique$`)
	if matches := pattern4.FindStringSubmatch(constraintName); len(matches) >= 2 {
		return matches[1]
	}

	// Pattern 5: uq_tablename_fieldname (e.g., uq_users_email)
	pattern5 := regexp.MustCompile(`^uq_[a-z]+_(.+)$`)
	if matches := pattern5.FindStringSubmatch(constraintName); len(matches) >= 2 {
		return matches[1]
	}

	// If no pattern matches, try to extract the last part after underscore
	parts := strings.Split(constraintName, "_")
	if len(parts) > 1 {
		// Remove common prefixes/suffixes
		lastPart := parts[len(parts)-1]
		if lastPart != "key" && lastPart != "constraint" && lastPart != "index" {
			return lastPart
		}
	}

	return ""
}

// Maps database errors to appropriate HTTP status codes
func determineStatusCodeFromError(err error) int {
	errMsg := err.Error()

	// Unique violation (duplicate key) → 409 Conflict
	if strings.Contains(errMsg, "23505") || strings.Contains(errMsg, "unique constraint") {
		return 409
	}

	// Foreign key violation → 400 Bad Request
	if strings.Contains(errMsg, "23503") || strings.Contains(errMsg, "foreign key") {
		return 400
	}

	// Not null violation → 400 Bad Request
	if strings.Contains(errMsg, "23502") || strings.Contains(errMsg, "null value") {
		return 400
	}

	// Invalid text representation → 400 Bad Request
	if strings.Contains(errMsg, "22P02") {
		return 400
	}

	// Connection errors → 503 Service Unavailable
	if strings.Contains(errMsg, "08001") || strings.Contains(errMsg, "connection") {
		return 503
	}

	// Authentication errors → 500 Internal Server Error
	if strings.Contains(errMsg, "28000") {
		return 500
	}

	// Default to 500 for unhandled database errors
	return 500
}
