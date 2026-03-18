# Validation Framework

Custom fluent validation API in `pkg/validator/`. Used in every service method to validate request structs before processing.

## Basic pattern

Create a validator, chain `Check()` calls for each field, then call `Error()` to get accumulated errors:

```go
func (req *CreateVendorRequest) Validate() error {
	v := validator.New()

	v.Check(req.OrganizationID, "organization_id",
		validator.Required(),
		validator.GID(coredata.OrganizationEntityType),
	)
	v.Check(req.Name, "name",
		validator.Required(),
		validator.SafeTextNoNewLine(TitleMaxLength),
	)
	v.Check(req.Category, "category",
		validator.OneOfSlice(coredata.VendorCategories()),
	)

	return v.Error()
}
```

`Check(value, fieldName, validators...)` runs validators sequentially on a value. Multiple `Check()` calls accumulate all errors. `Error()` returns `nil` if clean, or `ValidationErrors` (which implements `error`).

## Available validators

### Common
- `Required()` — value must not be nil, empty string, or empty slice
- `NotEmpty()` — value cannot be empty/nil (only checks content, not presence)

### String
- `MinLen(n)` — at least n characters
- `MaxLen(n)` — at most n characters
- `OneOfSlice[T](allowed)` — value must be in allowed list

### Numeric
- `Min(n)` — value >= n
- `Max(n)` — value <= n

### Format
- `URL()` — valid HTTP/HTTPS URL with host
- `HTTPSUrl()` — HTTPS-only URL
- `Domain()` — valid RFC-compliant domain name
- `GID(entityTypes...)` — valid GID, optionally restricted to specific entity types

### Security
- `NoHTML()` — rejects HTML tags
- `PrintableText()` — rejects invisible/harmful Unicode (control chars, bidi overrides, zero-width)
- `NoNewLine()` — rejects `\n` and `\r`
- `SafeText(maxLen)` — combines NotEmpty + MaxLen + NoHTML + PrintableText (allows newlines)
- `SafeTextNoNewLine(maxLen)` — same as SafeText but also rejects newlines (for single-line fields)

### Time
- `After(refTime)` — time must be after reference
- `Before(refTime)` — time must be before reference
- `RangeDuration(min, max)` — duration between min and max inclusive

## Pointer handling

The framework automatically dereferences pointers at any level. Nil pointers pass all non-`Required` validators:

```go
v.Check(stringValue, "field", validator.Required())    // Direct value
v.Check(&stringValue, "field", validator.Required())   // Pointer — auto-dereferenced
v.Check(nilPointer, "field", validator.MinLen(5))      // Nil passes (not Required)
v.Check(nilPointer, "field", validator.Required())     // Nil fails Required
```

## Collection validation

Use `CheckEach` to validate each item in a slice:

```go
v.CheckEach(ids, "ids", func(index int, item any) {
	gidValue := item.(gid.GID)
	v.Check(gidValue, fmt.Sprintf("ids[%d]", index),
		validator.Required(),
		validator.GID(coredata.VendorEntityType),
	)
})
```

Nil or empty slices are silently skipped.

## Error types

```go
type ValidationError struct {
	Field   string    // e.g. "email"
	Code    ErrorCode // e.g. ErrorCodeRequired
	Message string    // human-readable
	Value   any       // the problematic value
}
```

Error codes:

| Code | Meaning |
|------|---------|
| `REQUIRED` | Field is missing or empty |
| `INVALID_FORMAT` | Value does not match expected format |
| `OUT_OF_RANGE` | Numeric value outside bounds |
| `TOO_SHORT` | String below minimum length |
| `TOO_LONG` | String above maximum length |
| `INVALID_EMAIL` | Invalid email address |
| `INVALID_URL` | Invalid URL |
| `INVALID_ENUM` | Value not in allowed set |
| `INVALID_GID` | Invalid GID or wrong entity type |
| `UNSAFE_CONTENT` | HTML, control chars, or harmful Unicode |
| `CUSTOM` | Custom validation error |

`ValidationErrors` is a slice with query methods:

```go
errs := err.(validator.ValidationErrors)
errs.HasErrors()           // bool
errs.Fields()              // unique field names
errs.ByField("name")       // filter by field
errs.ByCode(ErrorCodeRequired) // filter by code
errs.First()               // first error
```

## Error propagation

Validation errors flow naturally through Go's error interface:

1. Request struct's `Validate()` returns `ValidationErrors` or `nil`
2. Service method checks error before processing
3. GraphQL/HTTP handlers convert `ValidationErrors` to appropriate response format

```go
func (s *Service) CreateVendor(ctx context.Context, req CreateVendorRequest) (*coredata.Vendor, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// proceed with business logic
}
```
