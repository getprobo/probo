# Go Testing

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
