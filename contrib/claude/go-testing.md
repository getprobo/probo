# Go Testing

Test library: `github.com/stretchr/testify` (`require` for fatal, `assert` for non-fatal).

## Prefer end-to-end tests

Prefer an end-to-end test over a unit test. End-to-end tests run against a live server and a real database, so they cover the path the product actually takes.

A unit test that creates objects in the database should usually be an end-to-end test instead. Persistence, authorization, and API behavior belong in the end-to-end suite. Keep unit tests for pure logic that takes plain values and returns a result, such as formatting and decisions. If the test needs rows in the database to set up its subject, write an end-to-end test. See [`e2e.md`](e2e.md).

## Package naming

Black-box test packages (`package foo_test`). White-box (`package foo`) only when testing unexported functions.

## Test naming

Test names follow `TestFunctionName_Scenario`. Use `t.Run` for subtests with lowercase descriptive names. Table-driven tests for parameterized cases.

## Parallel tests

Always call `t.Parallel()` at both the top-level test and each subtest.

## `require` vs `assert`

- `require` — stops the test immediately on failure. Use for **preconditions** that would make subsequent assertions meaningless: `require.NoError`, `require.Error`, `require.ErrorAs`, `require.Len`, `require.NotNil`, `require.True` (as a guard).
- `assert` — logs failure but continues the test. Use for the **actual values** being verified: `assert.Equal`, `assert.Contains`, `assert.True`, `assert.False`, `assert.Nil`.

Rule of thumb: if a failure would cause a nil-pointer panic or make every following assertion nonsensical, use `require`; otherwise use `assert`.

```go
func TestRun_Handoff(t *testing.T) {
	t.Parallel()

	t.Run(
		"handoff with custom tool name and description",
		func(t *testing.T) {
			t.Parallel()

			// ... setup ...

			result, err := triage.Run(
				context.Background(),
				[]llm.Message{
					userMessage("How much is my invoice?"),
				},
			)

			require.NoError(t, err)
			assert.Equal(t, "Your invoice is $42.", result.FinalMessage().Text())
			assert.Equal(t, "billing", result.LastAgent.Name())
		},
	)
}
```

## Helpers and mocks

Define mock types and helper functions (e.g. `stopResponse`, `userMessage`) at the top of the test file, not inline in each test.

Do not add a production interface so a test can mock it. If the behavior under test is formatting or a decision, call that function with plain values. Do not stand up fakes for PDF rendering or object storage from the caller.
