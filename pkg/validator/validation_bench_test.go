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

func BenchmarkValidate_NestedStruct(b *testing.B) {
	type Address struct {
		City    string
		ZipCode string
	}

	type User struct {
		Name    string
		Address Address
	}

	user := User{
		Name: "John Doe",
		Address: Address{
			City:    "New York",
			ZipCode: "10001",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := New()
		v.Check(&user.Name, "name", Required())
		v.CheckNested("address", func(av *Validator) {
			av.Check(&user.Address.City, "city", Required())
			av.Check(&user.Address.ZipCode, "zipCode", Pattern(`^\d{5}$`, ""))
		})
	}
}

func BenchmarkValidate_ArrayValidation(b *testing.B) {
	type Item struct {
		Name  string
		Price int
	}

	items := []Item{
		{Name: "Item 1", Price: 100},
		{Name: "Item 2", Price: 200},
		{Name: "Item 3", Price: 300},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := New()
		for j, item := range items {
			v.CheckNested("items[0]", func(iv *Validator) {
				iv.Check(&item.Name, "name", Required())
				iv.Check(&item.Price, "price", Min(0))
				_ = j
			})
		}
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

func BenchmarkUUID(b *testing.B) {
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	validator := UUID()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&uuid)
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

func BenchmarkMinFloat(b *testing.B) {
	num := 99.99
	validator := MinFloat(0.01)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&num)
	}
}

func BenchmarkMaxFloat(b *testing.B) {
	num := 50.50
	validator := MaxFloat(99.99)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&num)
	}
}

func BenchmarkRangeFloat(b *testing.B) {
	num := 50.50
	validator := RangeFloat(0.01, 99.99)

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

func BenchmarkPattern(b *testing.B) {
	zipCode := "12345"
	validator := Pattern(`^\d{5}$`, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&zipCode)
	}
}

func BenchmarkValidate_WithErrors(b *testing.B) {
	email := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := New()
		v.Check(&email, "email", Required(), NotEmpty())
		if !v.HasErrors() {
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
		v.Check(&user.Price, "price", MinFloat(0.01))

		v.CheckNested("address", func(av *Validator) {
			av.Check(&user.Address.Street, "street", Required())
			av.Check(&user.Address.City, "city", Required())
			av.Check(&user.Address.ZipCode, "zipCode", Pattern(`^\d{5}$`, ""))
		})
	}
}

func BenchmarkMinItems(b *testing.B) {
	items := []string{"a", "b", "c"}
	validator := MinItems(2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&items)
	}
}

func BenchmarkMaxItems(b *testing.B) {
	items := []string{"a", "b", "c"}
	validator := MaxItems(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&items)
	}
}

func BenchmarkUniqueItems(b *testing.B) {
	items := []string{"a", "b", "c"}
	validator := UniqueItems()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&items)
	}
}

func BenchmarkAlphaNumeric(b *testing.B) {
	str := "abc123DEF456"
	validator := AlphaNumeric()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str)
	}
}

func BenchmarkNoSpaces(b *testing.B) {
	str := "hello-world-test"
	validator := NoSpaces()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str)
	}
}

func BenchmarkSlug(b *testing.B) {
	str := "hello-world-123"
	validator := Slug()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str)
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

func BenchmarkFutureDate(b *testing.B) {
	future := time.Now().Add(24 * time.Hour)
	validator := FutureDate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&future)
	}
}

func BenchmarkPastDate(b *testing.B) {
	past := time.Now().Add(-24 * time.Hour)
	validator := PastDate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&past)
	}
}

func BenchmarkEqualTo(b *testing.B) {
	str1 := "password"
	str2 := "password"
	validator := EqualTo(&str2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str1)
	}
}

func BenchmarkNotEqualTo(b *testing.B) {
	str1 := "password"
	str2 := "different"
	validator := NotEqualTo(&str2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator(&str1)
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

func BenchmarkHTTPUrl(b *testing.B) {
	str := "http://api.example.com/v1/users"
	validator := HTTPUrl()

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
