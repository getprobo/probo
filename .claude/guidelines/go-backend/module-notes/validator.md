# Probo — Go Backend — pkg/validator

**Purpose.** Fluent in-memory validation framework used by every
mutating service request. No I/O.

**Key files.**

- `errors.go` — `Validator`, `ValidationError`, `ValidationErrors`
  (named `[]*ValidationError` implementing `error`), `ErrorCode`
  string constants (`REQUIRED`, `INVALID_FORMAT`, `OUT_OF_RANGE`, ...).
- Built-in validators: `Required`, `NotEmpty`, `GID(entityTypes...)`,
  `SafeText`, `SafeTextNoNewLine`, `MinLen`, `MaxLen`, `Min`, `Max`,
  `OneOfSlice`, `URL`, `HTTPSUrl`, `Domain`, `NoHTML`, `PrintableText`,
  `NoNewLine`, `After`, `Before`, `RangeDuration`.

**How to use (canonical).**

```go
func (r CreateVendorRequest) Validate() error {
    v := validator.New()
    v.Check(r.Name, "name", validator.Required(), validator.MaxLen(NameMaxLength))
    v.Check(r.Description, "description", validator.MaxLen(DescriptionMaxLength))
    v.Check(r.Email, "email", validator.Required(), validator.Email())
    return v.Error() // nil or ValidationErrors
}
```

**How to extend.** Add a new built-in by writing a `func(value any)
*ValidationError` and exposing it as `func XXX() CheckFunc { return ... }`.
Add a new `ErrorCode` constant if the failure category is genuinely
new — otherwise reuse `INVALID_FORMAT` or `CUSTOM`.

**Top pitfalls.**

- Re-using a `Validator` across calls. `New()` allocates fresh state;
  the Validator is per-call.
- Storing the offending value in `ValidationError.Value` and then
  logging the entire `ValidationErrors` payload — that may include PII
  (emails, names). Log only `Field` and `Code`; surface `Message` and
  `Value` to the user via the API response, not to logs.
