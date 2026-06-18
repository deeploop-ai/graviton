package crud

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// AscendingOrder represents ascending sort order (AIP-132)
	AscendingOrder = "asc"

	// DescendingOrder represents descending sort order (AIP-132)
	DescendingOrder = "desc"
)

// OrderByClause represents a parsed order by clause following AIP-132
// See: https://google.aip.dev/132#ordering
type OrderByClause struct {
	// Field is the field name to sort by
	Field string `json:"field"`

	// Direction is the sort direction ("asc" or "desc")
	Direction string `json:"direction"`
}

// ParseOrderBy parses an order_by string following AIP-132
//
// Examples:
//
//	"name" -> [{Field: "name", Direction: "asc"}]
//	"name desc" -> [{Field: "name", Direction: "desc"}]
//	"created_at asc, name desc" -> [{Field: "created_at", Direction: "asc"}, {Field: "name", Direction: "desc"}]
func ParseOrderBy(orderBy string) ([]OrderByClause, error) {
	if orderBy == "" {
		return nil, nil
	}

	// Split by comma for multiple fields
	parts := strings.Split(orderBy, ",")
	clauses := make([]OrderByClause, 0, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split field and direction
		components := strings.Fields(part)
		if len(components) == 0 {
			continue
		}

		clause := OrderByClause{
			Field:     components[0],
			Direction: AscendingOrder, // Default is ascending
		}

		if len(components) > 1 {
			direction := strings.ToLower(components[1])
			if direction != AscendingOrder && direction != DescendingOrder {
				return nil, fmt.Errorf("invalid sort direction '%s' at position %d, must be 'asc' or 'desc'", direction, i+1)
			}
			clause.Direction = direction
		}

		// Validate field name (alphanumeric, underscore, dot)
		if !isValidFieldName(clause.Field) {
			return nil, fmt.Errorf("invalid field name '%s' at position %d", clause.Field, i+1)
		}

		clauses = append(clauses, clause)
	}

	return clauses, nil
}

// isValidFieldName validates a field name
func isValidFieldName(field string) bool {
	if field == "" {
		return false
	}

	// Allow alphanumeric, underscore, and dot for nested fields
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_.]*$`, field)
	return matched
}

// ValidateOrderByFields validates that order_by fields are in the allowed list
func ValidateOrderByFields(clauses []OrderByClause, allowedFields map[string]bool) error {
	for _, clause := range clauses {
		if !allowedFields[clause.Field] {
			return fmt.Errorf("field '%s' is not allowed for ordering", clause.Field)
		}
	}
	return nil
}

// BuildOrderByClause builds an order by clause for SQL queries
func BuildOrderByClause(clauses []OrderByClause, fieldMappings map[string]string) string {
	if len(clauses) == 0 {
		return ""
	}

	var parts []string
	for _, clause := range clauses {
		// Use mapped field name if provided, otherwise use original field name
		fieldName := clause.Field
		if mapping, ok := fieldMappings[clause.Field]; ok {
			fieldName = mapping
		}

		// Defence-in-depth: skip unknown fragments unless trusted server-built SQL.
		if !IsTrustedSQLMappingFragment(fieldName) && !isValidFieldName(fieldName) {
			continue
		}

		direction := strings.ToUpper(clause.Direction)
		parts = append(parts, fmt.Sprintf("%s %s", fieldName, direction))
	}

	return strings.Join(parts, ", ")
}

// OrderByToString converts order by clauses back to string representation
func OrderByToString(clauses []OrderByClause) string {
	if len(clauses) == 0 {
		return ""
	}

	var parts []string
	for _, clause := range clauses {
		if clause.Direction == AscendingOrder {
			parts = append(parts, clause.Field)
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", clause.Field, clause.Direction))
		}
	}

	return strings.Join(parts, ", ")
}

// GetFieldDirection returns the direction for a specific field
func GetFieldDirection(clauses []OrderByClause, field string) string {
	for _, clause := range clauses {
		if clause.Field == field {
			return clause.Direction
		}
	}
	return AscendingOrder
}

// HasFieldDirection checks if a field has a specific direction
func HasFieldDirection(clauses []OrderByClause, field, direction string) bool {
	for _, clause := range clauses {
		if clause.Field == field && clause.Direction == direction {
			return true
		}
	}
	return false
}

// IsFieldOrdered checks if a field is in the order by clauses
func IsFieldOrdered(clauses []OrderByClause, field string) bool {
	for _, clause := range clauses {
		if clause.Field == field {
			return true
		}
	}
	return false
}

// PrimaryField returns the primary (first) sort field
func PrimaryField(clauses []OrderByClause) string {
	if len(clauses) == 0 {
		return ""
	}
	return clauses[0].Field
}

// PrimaryDirection returns the primary (first) sort direction
func PrimaryDirection(clauses []OrderByClause) string {
	if len(clauses) == 0 {
		return AscendingOrder
	}
	return clauses[0].Direction
}

// ReverseDirection reverses the direction of all clauses
func ReverseDirection(clauses []OrderByClause) []OrderByClause {
	result := make([]OrderByClause, len(clauses))
	for i, clause := range clauses {
		result[i] = OrderByClause{
			Field:     clause.Field,
			Direction: reverseDirection(clause.Direction),
		}
	}
	return result
}

// reverseDirection reverses a single direction
func reverseDirection(direction string) string {
	if direction == AscendingOrder {
		return DescendingOrder
	}
	return AscendingOrder
}

// MergeOrderBy merges multiple order by clause sets, later ones take precedence
func MergeOrderBy(sets ...[]OrderByClause) []OrderByClause {
	seen := make(map[string]bool)
	var result []OrderByClause

	// Process in reverse order so later sets take precedence
	for i := len(sets) - 1; i >= 0; i-- {
		for _, clause := range sets[i] {
			if !seen[clause.Field] {
				seen[clause.Field] = true
				result = append([]OrderByClause{clause}, result...)
			}
		}
	}

	return result
}

// DefaultOrderBy provides default ordering when no order_by is specified
func DefaultOrderBy(field string, direction string) []OrderByClause {
	if direction == "" {
		direction = AscendingOrder
	}
	return []OrderByClause{
		{Field: field, Direction: direction},
	}
}

// OrderByValidator validates order by clauses against allowed fields
type OrderByValidator struct {
	allowedFields   map[string]bool
	fieldMappings   map[string]string
	defaultOrder    []OrderByClause
	caseInsensitive bool
}

// NewOrderByValidator creates a new order by validator
func NewOrderByValidator() *OrderByValidator {
	return &OrderByValidator{
		allowedFields:   make(map[string]bool),
		fieldMappings:   make(map[string]string),
		caseInsensitive: false,
	}
}

// AddField adds an allowed field
func (v *OrderByValidator) AddField(field string) *OrderByValidator {
	v.allowedFields[field] = true
	return v
}

// AddFields adds multiple allowed fields
func (v *OrderByValidator) AddFields(fields ...string) *OrderByValidator {
	for _, field := range fields {
		v.allowedFields[field] = true
	}
	return v
}

// AddMapping adds a field mapping (proto field -> database field)
func (v *OrderByValidator) AddMapping(protoField, dbField string) *OrderByValidator {
	v.fieldMappings[protoField] = dbField
	v.allowedFields[protoField] = true
	return v
}

// SetDefault sets the default order by clause
func (v *OrderByValidator) SetDefault(clauses ...OrderByClause) *OrderByValidator {
	v.defaultOrder = clauses
	return v
}

// SetCaseInsensitive enables or disables case-insensitive field matching
func (v *OrderByValidator) SetCaseInsensitive(enabled bool) *OrderByValidator {
	v.caseInsensitive = enabled
	return v
}

// Parse validates and parses order by string
func (v *OrderByValidator) Parse(orderBy string) ([]OrderByClause, error) {
	// If empty and default is set, use default
	if orderBy == "" && len(v.defaultOrder) > 0 {
		return v.defaultOrder, nil
	}

	// Parse the order by string
	clauses, err := ParseOrderBy(orderBy)
	if err != nil {
		return nil, err
	}

	// If empty after parsing and default is set, use default
	if len(clauses) == 0 && len(v.defaultOrder) > 0 {
		return v.defaultOrder, nil
	}

	// Validate fields are allowed
	if err := v.Validate(clauses); err != nil {
		return nil, err
	}

	return clauses, nil
}

// Validate validates order by clauses
func (v *OrderByValidator) Validate(clauses []OrderByClause) error {
	return ValidateOrderByFields(clauses, v.allowedFields)
}

// BuildSQL builds SQL order by clause with field mappings
func (v *OrderByValidator) BuildSQL(clauses []OrderByClause) string {
	return BuildOrderByClause(clauses, v.fieldMappings)
}

// OrderByOptions provides configuration for order by parsing
type OrderByOptions struct {
	AllowedFields    []string
	FieldMappings    map[string]string
	DefaultField     string
	DefaultDirection string
	CaseInsensitive  bool
}

// ParseOrderByWithOptions parses order by with options
func ParseOrderByWithOptions(orderBy string, opts OrderByOptions) ([]OrderByClause, error) {
	// Build validator
	validator := NewOrderByValidator()
	validator.AddFields(opts.AllowedFields...)

	for protoField, dbField := range opts.FieldMappings {
		validator.AddMapping(protoField, dbField)
	}

	if opts.DefaultField != "" {
		direction := opts.DefaultDirection
		if direction == "" {
			direction = AscendingOrder
		}
		validator.SetDefault(OrderByClause{
			Field:     opts.DefaultField,
			Direction: direction,
		})
	}

	validator.SetCaseInsensitive(opts.CaseInsensitive)

	// Parse and validate
	return validator.Parse(orderBy)
}

// SafeOrderBy safely parses order by, returning default on error
func SafeOrderBy(orderBy string, defaultField string) []OrderByClause {
	clauses, err := ParseOrderBy(orderBy)
	if err != nil || len(clauses) == 0 {
		return DefaultOrderBy(defaultField, AscendingOrder)
	}
	return clauses
}
