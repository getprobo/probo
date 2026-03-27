# Go Style

Layout and readability rules for Go source. (Error handling, naming, and imports are covered in `AGENTS.md` / other guides.)

## Call expressions and argument lists

In the [Go spec](https://go.dev/ref/spec#Calls), a **call** is a primary expression `f(a1, a2, … an)` where `f` is the **function value** (or **method value**) and `a1` … `an` are **arguments** passed to the matching parameters.

Treat the **argument list** as either single-line or multiline — never mixed:

- **Single-line call** — the entire call, from the callee through the closing `)`, fits on one source line. Any argument may be a short expression (including a one-line composite literal or conversion).
- **Multiline call** — if any argument is written across multiple lines (e.g. a multi-line **composite literal**, **function literal**, or other expression that contains a line break), then **every** argument must start on its own line: one argument per line at the top level of that argument list. The closing `)` is on its own line after the last argument (with a trailing comma after the final argument when the list is multiline).

Do not place some arguments on the same line as the opening `(` while others continue on following lines.

```go
// Good — entire call on one line
id := gid.New(tenantID, "Foo")

// Good — multiline argument list; each argument on its own line
svc, err := foo.NewService(
	ctx,
	db,
	logger,
	foo.Config{
		Interval: 10 * time.Second,
		MaxRetry: 3,
	},
)

// Good — function literal argument is multiline, so the name argument is on its own line too
t.Run(
	"handoff with custom tool name",
	func(t *testing.T) {
		t.Parallel()
		// ...
	},
)

// Bad — mixed: first arguments on the callee line, last argument is a multiline composite literal
svc, err := foo.NewService(ctx, db, logger, foo.Config{
	Interval: 10 * time.Second,
})
```

The same rule applies to **method calls** `x.M(a1, …)` — the receiver is already bound; the rule applies to the **argument list** after the method name.
