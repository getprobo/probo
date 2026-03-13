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
	"testing"
	"time"
)

func BenchmarkValidate_SingleField(b *testing.B) {
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := New()
		v.Check(&email, "email", Required(), NotEmpty())
	}
}

func BenchmarkValidate_MultipleFields(b *testing.B) {
	email := "test@example.com"
	password := "password123"
	age := 25

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := New()
		v.Check(&email, "email", Required(), NotEmpty())
		v.Check(&password, "password", Required(), MinLen(8))
		v.Check(&age, "age", Min(18), Max(120))
	}
}

func BenchmarkValidate_OptionalField(b *testing.B) {
	var website *string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := New()
		v.Check(website, "website", URL())
	}
}

func BenchmarkURL(b *testing.B) {
	urlStr := "https://example.com"
	validator := URL()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&urlStr)
	}
}

func BenchmarkMinLen(b *testing.B) {
	str := "hello world"
	validator := MinLen(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str)
	}
}

func BenchmarkMin(b *testing.B) {
	num := 42
	validator := Min(18)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&num)
	}
}

func BenchmarkNotEmpty(b *testing.B) {
	str := "hello world"
	validator := NotEmpty()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str)
	}
}

func BenchmarkValidate_WithErrors(b *testing.B) {
	email := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := New()
		v.Check(&email, "email", Required(), NotEmpty())
		if v.Error() == nil {
			b.Fatal("expected validation error")
		}
	}
}

func BenchmarkValidate_ComplexForm(b *testing.B) {
	type Address struct {
		Street  string
		City    string
		ZipCode string
	}

	type User struct {
		Email       string
		Name        string
		Age         int
		Website     *string
		PhoneNumber *string
		Price       float64
		Address     Address
	}

	website := "https://example.com"
	user := User{
		Email:       "user@example.com",
		Name:        "John Doe",
		Age:         30,
		Website:     &website,
		PhoneNumber: nil,
		Price:       99.99,
		Address: Address{
			Street:  "123 Main St",
			City:    "New York",
			ZipCode: "10001",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := New()
		v.Check(&user.Email, "email", Required(), NotEmpty())
		v.Check(&user.Name, "name", Required(), MinLen(2))
		v.Check(&user.Age, "age", Min(18), Max(120))
		v.Check(user.Website, "website", URL())
		v.Check(user.PhoneNumber, "phoneNumber", MinLen(10))
	}
}

func BenchmarkAfter(b *testing.B) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	validator := After(now)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&future)
	}
}

func BenchmarkBefore(b *testing.B) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	validator := Before(now)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&past)
	}
}

func BenchmarkDomain(b *testing.B) {
	str := "api.example.com"
	validator := Domain()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str)
	}
}

func BenchmarkHTTPSUrl(b *testing.B) {
	str := "https://api.example.com/v1/users"
	validator := HTTPSUrl()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str)
	}
}
