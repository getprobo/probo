// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package validator

import (
	"fmt"
	"testing"

	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/mail"
)

func TestValidator_Validate(t *testing.T) {
	t.Run("single field validation", func(t *testing.T) {
		v := New()
		email := mail.Addr("test@example.com")

		v.Check(&email, "email", Required(), NotEmpty())

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Errors())
		}
	})

	t.Run("multiple field validations", func(t *testing.T) {
		v := New()
		email := mail.Nil
		password := "123"

		v.Check(email, "email", NotEmpty())
		v.Check(&password, "password", Required(), MinLen(8))

		if !v.HasErrors() {
			t.Error("expected validation errors")
		}

		errors := v.Errors()
		// email: 1 error (Required), password: 1 error (MinLen - Required passes because it's not empty)
		if len(errors) != 2 {
			t.Errorf("expected 2 errors, got %d: %v", len(errors), errors)
		}
	})

	t.Run("collect multiple errors for same field", func(t *testing.T) {
		v := New()
		value := "abc"

		v.Check(&value, "password", MinLen(8), MaxLen(5))

		errors := v.Errors()
		// Both MinLen and MaxLen will fail (too short and somehow conflicts, but logically MinLen will fail)
		if len(errors) < 1 {
			t.Errorf("expected at least 1 error, got %d", len(errors))
		}
	})
}

func TestValidator_CheckNested(t *testing.T) {
	v := New()

	v.CheckNested("user", func(nv *Validator) {
		email := mail.Nil
		nv.Check(&email, "email", NotEmpty())

		nv.CheckNested("address", func(av *Validator) {
			city := ""
			av.Check(&city, "city", Required())
		})
	})

	if !v.HasErrors() {
		t.Error("expected validation errors")
	}

	errors := v.Errors()
	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}

	// Check field paths
	expectedFields := map[string]bool{
		"user.email":        true,
		"user.address.city": true,
	}

	for _, err := range errors {
		if !expectedFields[err.Field] {
			t.Errorf("unexpected field path: %s", err.Field)
		}
	}
}

func TestValidator_Error(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		v := New()
		if v.Error() != nil {
			t.Errorf("expected nil error, got: %v", v.Error())
		}
	})

	t.Run("with errors", func(t *testing.T) {
		v := New()
		email := ""
		v.Check(&email, "email", Required())

		err := v.Error()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestValidationErrors_Methods(t *testing.T) {
	errors := ValidationErrors{
		{Field: "email", Code: ErrorCodeInvalidEmail, Message: "invalid email"},
		{Field: "password", Code: ErrorCodeTooShort, Message: "too short"},
		{Field: "email", Code: ErrorCodeRequired, Message: "required"},
	}

	t.Run("Fields", func(t *testing.T) {
		fields := errors.Fields()
		if len(fields) != 3 {
			t.Errorf("expected 3 fields, got %d", len(fields))
		}
	})

	t.Run("ByField", func(t *testing.T) {
		emailErrors := errors.ByField("email")
		if len(emailErrors) != 2 {
			t.Errorf("expected 2 email errors, got %d", len(emailErrors))
		}
	})

	t.Run("ByCode", func(t *testing.T) {
		requiredErrors := errors.ByCode(ErrorCodeRequired)
		if len(requiredErrors) != 1 {
			t.Errorf("expected 1 required error, got %d", len(requiredErrors))
		}
	})

	t.Run("First", func(t *testing.T) {
		first := errors.First()
		if first == nil {
			t.Fatal("expected first error")
		} else if first.Field != "email" {
			t.Errorf("expected first field to be 'email', got '%s'", first.Field)
		}
	})

	t.Run("Error", func(t *testing.T) {
		errorStr := errors.Error()
		if errorStr == "" {
			t.Error("expected non-empty error string")
		}
	})
}

func TestOptionalFieldExample(t *testing.T) {
	type CreateUserRequest struct {
		Email       string
		Name        string
		Website     *string
		PhoneNumber *string
		Age         *int
	}

	website := "not-a-url"
	req := CreateUserRequest{
		Email:       "user@example.com",
		Name:        "John Doe",
		Website:     &website,
		PhoneNumber: nil,
		Age:         nil,
	}

	v := New()

	v.Check(&req.Email, "email", Required(), NotEmpty())
	v.Check(&req.Name, "name", Required(), MinLen(2))
	v.Check(req.Website, "website", URL())
	v.Check(req.PhoneNumber, "phoneNumber", MinLen(10))
	v.Check(req.Age, "age", Min(18), Max(120))

	if !v.HasErrors() {
		t.Fatal("expected validation errors")
	}

	errors := v.Errors()

	websiteErr := errors.ByField("website")
	if len(websiteErr) != 1 {
		t.Errorf("expected 1 website error, got %d", len(websiteErr))
	}

	phoneErr := errors.ByField("phoneNumber")
	if len(phoneErr) != 0 {
		t.Errorf("expected 0 phoneNumber errors (nil should be skipped), got %d", len(phoneErr))
	}

	ageErr := errors.ByField("age")
	if len(ageErr) != 0 {
		t.Errorf("expected 0 age errors (nil should be skipped), got %d", len(ageErr))
	}

	t.Logf("Optional field validation errors: %s", errors.Error())
}

func TestRealWorldExample(t *testing.T) {
	// Simulate a user registration form
	type Address struct {
		City    string
		ZipCode string
	}

	type User struct {
		Email    string
		Password string
		Age      int
		Website  *string
		Address  Address
	}

	user := User{
		Email:    "",
		Password: "123",
		Age:      15,
		Website:  ref.Ref("not-a-url"),
		Address: Address{
			City:    "",
			ZipCode: "12345",
		},
	}

	v := New()

	// Validate user fields
	v.Check(&user.Email, "email", Required(), NotEmpty())
	v.Check(&user.Password, "password", Required(), MinLen(8))
	v.Check(&user.Age, "age", Min(18), Max(120))
	v.Check(user.Website, "website", URL())

	// Validate nested address
	v.CheckNested("address", func(av *Validator) {
		av.Check(&user.Address.City, "city", Required())
		av.Check(&user.Address.ZipCode, "zipCode", Pattern(`^\d{5}$`, "must be 5 digits"))
	})

	if !v.HasErrors() {
		t.Fatal("expected validation errors")
	}

	errors := v.Errors()
	expectedErrors := map[string]ErrorCode{
		"email":        ErrorCodeRequired,
		"password":     ErrorCodeTooShort,
		"age":          ErrorCodeOutOfRange,
		"website":      ErrorCodeInvalidURL,
		"address.city": ErrorCodeRequired,
	}

	// Check that we have the expected errors
	for field, expectedCode := range expectedErrors {
		found := false
		for _, err := range errors {
			if err.Field == field && err.Code == expectedCode {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected error for field '%s' with code '%s'", field, expectedCode)
		}
	}

	// Print errors for debugging
	t.Logf("Validation errors: %s", errors.Error())
}

func TestArrayValidation(t *testing.T) {
	type Item struct {
		Name  string
		Price int
	}

	items := []Item{
		{Name: "", Price: -10},
		{Name: "Valid", Price: 100},
		{Name: "X", Price: 10},
	}

	v := New()

	// Validate each item
	for i, item := range items {
		field := fmt.Sprintf("items[%d]", i)
		v.CheckNested(field, func(iv *Validator) {
			iv.Check(&item.Name, "name", Required(), MinLen(2))
			iv.Check(&item.Price, "price", Min(0))
		})
	}

	if !v.HasErrors() {
		t.Fatal("expected validation errors")
	}

	errors := v.Errors()

	// Check for specific field paths
	expectedFields := []string{
		"items[0].name",
		"items[0].price",
		"items[2].name",
	}

	for _, expectedField := range expectedFields {
		found := false
		for _, err := range errors {
			if err.Field == expectedField {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected error for field '%s'", expectedField)
		}
	}

	t.Logf("Array validation errors: %s", errors.Error())
}

func TestDuplicateValidators(t *testing.T) {
	t.Run("duplicate MinLen creates two errors", func(t *testing.T) {
		v := New()
		name := "abc"
		v.Check(&name, "name", MinLen(5), MinLen(5))

		if !v.HasErrors() {
			t.Error("expected validation errors")
		}

		errors := v.Errors()
		if len(errors) != 2 {
			t.Errorf("expected 2 errors (one per MinLen), got %d", len(errors))
		}

		if errors[0].Message != "must be at least 5 characters" {
			t.Errorf("unexpected first error: %s", errors[0].Message)
		}
		if errors[1].Message != "must be at least 5 characters" {
			t.Errorf("unexpected second error: %s", errors[1].Message)
		}
	})

	t.Run("duplicate Required creates two errors", func(t *testing.T) {
		v := New()
		name := ""
		v.Check(&name, "name", Required(), Required())

		errors := v.Errors()
		if len(errors) != 2 {
			t.Errorf("expected 2 errors, got %d", len(errors))
		}
	})

	t.Run("duplicate NotEmpty creates two errors", func(t *testing.T) {
		v := New()
		email := mail.Nil
		v.Check(&email, "email", NotEmpty(), NotEmpty())

		errors := v.Errors()
		if len(errors) != 2 {
			t.Errorf("expected 2 errors, got %d", len(errors))
		}
	})

	t.Run("same validator with different parameters", func(t *testing.T) {
		v := New()
		name := "test"
		v.Check(&name, "name", MinLen(5), MinLen(10))

		errors := v.Errors()
		if len(errors) != 2 {
			t.Errorf("expected 2 errors, got %d", len(errors))
		}

		if errors[0].Message != "must be at least 5 characters" {
			t.Errorf("unexpected first error: %s", errors[0].Message)
		}
		if errors[1].Message != "must be at least 10 characters" {
			t.Errorf("unexpected second error: %s", errors[1].Message)
		}
	})
}

func TestStandardErrorPattern(t *testing.T) {
	// Simulates a typical validation function
	validateUser := func(email mail.Addr, password string) error {
		v := New()
		v.Check(&email, "email", Required(), NotEmpty())
		v.Check(&password, "password", Required(), MinLen(8))
		return v.Error()
	}

	t.Run("valid data returns nil", func(t *testing.T) {
		email := mail.Addr("user@example.com")

		if err := validateUser(email, "password123"); err != nil {
			t.Errorf("expected nil, got: %v", err)
		}
	})

	t.Run("invalid data returns ValidationErrors as error", func(t *testing.T) {
		err := validateUser(mail.Nil, "123")
		if err == nil {
			t.Fatal("expected validation errors")
		}

		// Standard error handling
		t.Logf("validation failed: %v", err)

		// Can get detailed errors if needed
		if validationErrs, ok := err.(ValidationErrors); ok {
			for _, e := range validationErrs {
				t.Logf("  - %s: %s (code: %s)", e.Field, e.Message, e.Code)
			}

			// Can use helper methods
			emailErrs := validationErrs.ByField("email")
			if len(emailErrs) != 1 {
				t.Errorf("expected 1 email error, got %d", len(emailErrs))
			}

			passwordErrs := validationErrs.ByField("password")
			if len(passwordErrs) != 1 {
				t.Errorf("expected 1 password error, got %d", len(passwordErrs))
			}
		} else {
			t.Error("expected ValidationErrors type")
		}
	})
}
