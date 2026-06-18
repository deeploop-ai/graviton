package crud

import (
	"testing"
)

func TestFilterBuilder_Basic(t *testing.T) {
	// Test basic filter building
	filter := NewFilterBuilder().
		Equal("status", "active").
		And().
		GreaterThan("age", "18").
		Build()

	if filter == nil {
		t.Fatal("Expected filter to be non-nil")
	}

	if len(filter.Expressions) != 2 {
		t.Errorf("Expected 2 expressions, got %d", len(filter.Expressions))
	}

	// Check first expression
	expr1 := filter.Expressions[0]
	if expr1.Field != "status" {
		t.Errorf("Expected field 'status', got '%s'", expr1.Field)
	}
	if expr1.Operator != OperatorEqual {
		t.Errorf("Expected operator '=', got '%s'", expr1.Operator)
	}
	if expr1.Value != "active" {
		t.Errorf("Expected value 'active', got '%s'", expr1.Value)
	}

	// Check logical operator
	if expr1.LogicalOperator != LogicalAND {
		t.Errorf("Expected logical operator 'AND', got '%s'", expr1.LogicalOperator)
	}

	// Check second expression
	expr2 := filter.Expressions[1]
	if expr2.Field != "age" {
		t.Errorf("Expected field 'age', got '%s'", expr2.Field)
	}
	if expr2.Operator != OperatorGreaterThan {
		t.Errorf("Expected operator '>', got '%s'", expr2.Operator)
	}
}

func TestFilterBuilder_Or(t *testing.T) {
	filter := NewFilterBuilder().
		Equal("status", "active").
		Or().
		Equal("status", "pending").
		Build()

	if len(filter.Expressions) != 2 {
		t.Fatalf("Expected 2 expressions, got %d", len(filter.Expressions))
	}

	if filter.Expressions[0].LogicalOperator != LogicalOR {
		t.Errorf("Expected logical operator 'OR', got '%s'", filter.Expressions[0].LogicalOperator)
	}
}

func TestFilterBuilder_In(t *testing.T) {
	filter := NewFilterBuilder().
		In("status", "active", "pending", "suspended").
		Build()

	if len(filter.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(filter.Expressions))
	}

	expr := filter.Expressions[0]
	if expr.Operator != OperatorIn {
		t.Errorf("Expected operator 'in', got '%s'", expr.Operator)
	}
	if expr.Value != "active,pending,suspended" {
		t.Errorf("Expected value 'active,pending,suspended', got '%s'", expr.Value)
	}
}

func TestFilterBuilder_Contains(t *testing.T) {
	filter := NewFilterBuilder().
		Contains("name", "John").
		Build()

	if len(filter.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(filter.Expressions))
	}

	expr := filter.Expressions[0]
	if expr.Operator != OperatorContains {
		t.Errorf("Expected operator ':', got '%s'", expr.Operator)
	}
	if expr.Value != "John" {
		t.Errorf("Expected value 'John', got '%s'", expr.Value)
	}
}

func TestFilterBuilder_NotEqual(t *testing.T) {
	filter := NewFilterBuilder().
		NotEqual("status", "deleted").
		Build()

	if len(filter.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(filter.Expressions))
	}

	expr := filter.Expressions[0]
	if expr.Operator != OperatorNotEqual {
		t.Errorf("Expected operator '!=', got '%s'", expr.Operator)
	}
}

func TestFilterBuilder_ComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		builder  *FilterBuilder
		operator FilterOperator
	}{
		{
			name:     "LessThan",
			builder:  NewFilterBuilder().LessThan("age", "18"),
			operator: OperatorLessThan,
		},
		{
			name:     "LessThanOrEqual",
			builder:  NewFilterBuilder().LessThanOrEqual("age", "18"),
			operator: OperatorLessThanOrEqual,
		},
		{
			name:     "GreaterThan",
			builder:  NewFilterBuilder().GreaterThan("age", "18"),
			operator: OperatorGreaterThan,
		},
		{
			name:     "GreaterThanOrEqual",
			builder:  NewFilterBuilder().GreaterThanOrEqual("age", "18"),
			operator: OperatorGreaterThanOrEqual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.builder.Build()
			if len(filter.Expressions) != 1 {
				t.Fatalf("Expected 1 expression, got %d", len(filter.Expressions))
			}
			if filter.Expressions[0].Operator != tt.operator {
				t.Errorf("Expected operator '%s', got '%s'", tt.operator, filter.Expressions[0].Operator)
			}
		})
	}
}

func TestFilterBuilder_NotContains(t *testing.T) {
	filter := NewFilterBuilder().
		NotContains("name", "test").
		Build()

	if len(filter.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(filter.Expressions))
	}

	expr := filter.Expressions[0]
	if expr.Operator != OperatorNotContains {
		t.Errorf("Expected operator '!:', got '%s'", expr.Operator)
	}
}

func TestFilterBuilder_NotIn(t *testing.T) {
	filter := NewFilterBuilder().
		NotIn("status", "deleted", "banned").
		Build()

	if len(filter.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(filter.Expressions))
	}

	expr := filter.Expressions[0]
	if expr.Operator != OperatorNotIn {
		t.Errorf("Expected operator 'not_in', got '%s'", expr.Operator)
	}
}

func TestFilterBuilder_Has(t *testing.T) {
	filter := NewFilterBuilder().
		Has("metadata", "key").
		Build()

	if len(filter.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(filter.Expressions))
	}

	expr := filter.Expressions[0]
	if expr.Operator != OperatorHas {
		t.Errorf("Expected operator 'has', got '%s'", expr.Operator)
	}
}

func TestFilterBuilder_Expr(t *testing.T) {
	filter := NewFilterBuilder().
		Expr("status", OperatorEqual, "active").
		Build()

	if len(filter.Expressions) != 1 {
		t.Fatalf("Expected 1 expression, got %d", len(filter.Expressions))
	}

	expr := filter.Expressions[0]
	if expr.Field != "status" || expr.Operator != OperatorEqual || expr.Value != "active" {
		t.Errorf("Unexpected expression: %+v", expr)
	}
}

func TestFilterBuilder_ExprWithLogical(t *testing.T) {
	filter := NewFilterBuilder().
		ExprWithLogical("status", OperatorEqual, "active", LogicalAND).
		ExprWithLogical("age", OperatorGreaterThan, "18", "").
		Build()

	if len(filter.Expressions) != 2 {
		t.Fatalf("Expected 2 expressions, got %d", len(filter.Expressions))
	}

	if filter.Expressions[0].LogicalOperator != LogicalAND {
		t.Errorf("Expected logical operator 'AND', got '%s'", filter.Expressions[0].LogicalOperator)
	}
}

func TestFilterBuilder_String(t *testing.T) {
	filter := NewFilterBuilder().
		Equal("status", "active").
		And().
		GreaterThan("age", "18").
		Build()

	str := FilterToString(filter)
	expected := "status = active AND age > 18"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}

func TestFilterBuilder_MustBuild_Valid(t *testing.T) {
	filter := NewFilterBuilder().
		Equal("status", "active").
		MustBuild("status", "age")

	if filter == nil {
		t.Fatal("Expected filter to be non-nil")
	}
}

func TestFilterBuilder_MustBuild_Invalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when using invalid field, but did not panic")
		}
	}()

	NewFilterBuilder().
		Equal("invalid_field", "value").
		MustBuild("status", "age")
}

func TestFilterBuilder_BuildSQL(t *testing.T) {
	filter := NewFilterBuilder().
		Equal("status", "active").
		And().
		GreaterThan("age", "18").
		Build()

	fieldMappings := map[string]string{
		"status": "status",
		"age":    "user_age",
	}
	var args []interface{}
	where := BuildSQLWhere(filter, fieldMappings, &args)

	if where == "" {
		t.Error("Expected WHERE clause to be non-empty")
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

func TestNewFilterBuilder_Empty(t *testing.T) {
	filter := NewFilterBuilder().Build()

	if filter == nil {
		t.Fatal("Expected filter to be non-nil")
	}

	if len(filter.Expressions) != 0 {
		t.Errorf("Expected 0 expressions, got %d", len(filter.Expressions))
	}

	if !IsEmpty(filter) {
		t.Error("Expected empty filter")
	}
}
