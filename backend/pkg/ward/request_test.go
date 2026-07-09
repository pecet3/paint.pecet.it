package ward

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator"
)

type TestUser struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=18"`
}

func TestGetJson_Success(t *testing.T) {
	// Arrange: Prepare JSON data and the HTTP request
	jsonBody := `{"name": "John", "email": "john@example.com", "age": 25}`
	httpReq := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonBody))

	req := &Request{Http: httpReq}
	var u TestUser

	// Act: Call the method under test
	err := req.GetJson(&u)

	// Assert: Verify the results
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if u.Name != "John" || u.Email != "john@example.com" || u.Age != 25 {
		t.Errorf("Data was not parsed correctly: %+v", u)
	}
}

func TestGetJson_InvalidJSON(t *testing.T) {
	// Arrange: Malformed JSON (missing closing brace)
	jsonBody := `{"name": "John"`
	httpReq := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonBody))

	req := &Request{Http: httpReq}
	var u TestUser

	// Act
	err := req.GetJson(&u)

	// Assert
	if err == nil {
		t.Error("Expected a JSON parsing error, but got nil")
	}
}

func TestGetValidJson_Success(t *testing.T) {
	// Arrange: Structurally correct JSON that meets validation rules
	jsonBody := `{"name": "John", "email": "john@example.com", "age": 20}`
	httpReq := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonBody))

	req := &Request{Http: httpReq}
	var u TestUser

	// Act
	err := req.GetValidJson(&u)

	// Assert
	if err != nil {
		t.Fatalf("Expected no validation error, got: %v", err)
	}
}

func TestGetValidJson_ValidationError(t *testing.T) {
	// Arrange: Correct JSON syntax, but email is invalid and age is too low (< 18)
	jsonBody := `{"name": "John", "email": "invalid-email", "age": 15}`
	httpReq := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonBody))

	req := &Request{Http: httpReq}
	var u TestUser

	// Act
	err := req.GetValidJson(&u)

	// Assert
	if err == nil {
		t.Fatal("Expected a validation error, but got nil")
	}

	// Verify that the error is of type validator.ValidationErrors
	var valErrs validator.ValidationErrors
	if !errors.As(err, &valErrs) {
		t.Errorf("Expected error of type validator.ValidationErrors, got %T", err)
	}

	// Optional: Check the exact number of validation errors (should be 2: email and age)
	if len(valErrs) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(valErrs))
	}
}
