# Google Go Style Guide — Key Rules

> Distilled from https://google.github.io/styleguide/go/ (Guide + Decisions + Best Practices).
> Apply these rules when writing, reviewing, or refactoring Go code in this project.

---

## Principles (Priority Order)

1. **Clarity** — purpose and rationale obvious to readers
2. **Simplicity** — simplest approach that works
3. **Concision** — high signal-to-noise ratio
4. **Maintainability** — easy to modify correctly
5. **Consistency** — same patterns throughout the codebase

---

## Naming

### General Rules

- **MixedCaps** always. Never `snake_case` for identifiers. Filenames may use underscores.
- Name length proportional to scope: 1-letter for tiny scopes, multi-word for large scopes.
- Name by role, not by value or type.

### Packages

- Lowercase only, no underscores: `tabwriter` not `tab_writer`.
- Avoid generic names: `util`, `helper`, `common`.
- Consider how the package reads at call sites: `db.Load()` not `db.LoadFromDatabase()`.

### Exported Symbols

- Avoid repeating the package name: `widget.New()` not `widget.NewWidget()`.
- Getters: no `Get` prefix — `Counts()` not `GetCounts()`.
- Use `Compute`/`Fetch` to signal expensive or blocking operations.

### Receivers

- 1-2 letter abbreviation, consistent across all methods: `func (r *repository)`, `func (t Transaction)`.

### Constants

- MixedCaps, never `ALL_CAPS` or `K` prefix.
- Only create constants with semantic meaning — not for literal values.

### Initialisms

- All same case: `XMLAPI`, `htmlParser`, `HTTPClient`. Never `XmlApi` or `HttpClient`.

---

## Error Handling

### Returning Errors

- Always `error` as last return value.
- Return `nil` error for success.
- Exported functions return `error` type, not concrete error types.

### Error Strings

- Lowercase, no trailing punctuation: `"amount cannot be negative"`.
- Reason: error strings appear inside other context (`fmt.Errorf("create: %v", err)`).

### Wrapping: `%w` vs `%v`

| Use       | When                                                              |
|-----------|-------------------------------------------------------------------|
| `%w`      | Callers need `errors.Is`/`errors.As` — preserves the error chain |
| `%v`      | System boundaries (RPC, storage) — hides internals               |

Place `%w` at the end: `fmt.Errorf("fetch user: %w", err)`.

### Error Flow

Handle errors first, then proceed — no `else` after early return:

```go
val, err := doSomething()
if err != nil {
    return fmt.Errorf("do something: %w", err)
}
// happy path continues unindented
```

### Sentinel Errors

- Define with `errors.New`, check with `errors.Is`.
- Never match errors by string comparison against `.Error()` output.

### Don't Over-Annotate

Don't add context that the underlying error already provides:

```go
// Good — adds context the error lacks
return fmt.Errorf("launch codes unavailable: %v", err)

// Bad — duplicates the file path from os.Open
return fmt.Errorf("could not open settings.txt: %v", err)
```

---

## Declarations

### Variables

| Scenario         | Form                      |
|------------------|---------------------------|
| Non-zero value   | `i := 42`                 |
| Zero value       | `var coords Point`        |
| Pointer to zero  | `msg := new(pb.Message)`  |

### Slices

- Prefer nil: `var t []string` not `t := []string{}`.
- Check emptiness with `len(s) == 0`, never `s == nil`.
- No functional difference between nil and empty for most operations.

### Composite Literals

- Use field names in struct literals (especially for external packages).
- Omit zero-value fields when clarity is preserved.
- Run `gofmt -s` to simplify repeated type names in slice/map literals.

---

## Functions & Methods

### Design

- **Synchronous by default** — callers add concurrency when needed.
- Keep signatures on a single line; extract locals to shorten call sites.
- Don't let argument lists grow unbounded — use option structs or variadic options for complex configuration.

### Pass Values

- Pass by value unless mutation or large struct. Never pass pointer-to-string or pointer-to-interface.
- Exception: proto messages, large structs.

### Receiver Type

- **Pointer** when: mutates receiver, has unsafe-to-copy fields (`sync.Mutex`), or is large.
- **Value** when: receiver is small/immutable, or is a slice/map/func/channel.
- Prefer all methods on a type to use the same receiver kind.

### Panics

- Never for normal error handling — use `error` and multiple returns.
- Acceptable: `Must` functions at program startup, internal parser recovery (never escapes package).
- In `main`: prefer `log.Fatal`/`log.Exit` over `panic`.

---

## Interfaces

- Define in the **consumer** package, not the implementor.
- Implementors return concrete types (not interfaces).
- Don't define before realistic usage exists.
- Don't export interfaces that no external consumer uses.

---

## Comments & Documentation

### Doc Comments

- All exported names need doc comments.
- Full sentences starting with the symbol name: `// Parse reads the config file and returns...`
- Document: non-obvious concurrency semantics, cleanup requirements (`defer Close()`), significant error values.

### Internal Comments

- Explain *why*, not *what*.
- Signal deviations from standard patterns.
- Avoid redundant comments that restate the code.

### Comment Style

- Aim for ~80 columns but no hard limit.
- Long URLs may exceed the width for readability.

---

## Imports

### Grouping Order

1. Standard library
2. Project and third-party packages
3. Side-effect imports (`import _`)

### Rules

- Avoid renaming imports — only when collision or uninformative name.
- Never use dot imports (`import .`).
- Blank imports (`import _`) only in `main` or tests.

---

## Testing

### No Assertion Libraries

Use stdlib `testing` + `cmp.Equal`/`cmp.Diff`. No custom DSLs or assertion frameworks.

### Failure Messages

Format: `FuncName(input) = got, want expected`. Always got-before-want.

Include: function name, inputs, actual result, expected result.

### Table-Driven Tests

```go
tests := []struct {
    name string
    input  *Foo
    want   error
}{
    {"empty name", &Foo{}, ErrInvalid},
    {"valid", &Foo{Name: "ok"}, nil},
}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        _, err := ctrl.Create(tc.input)
        if !errors.Is(err, tc.want) {
            t.Errorf("Create(%v) error = %v, want %v", tc.input, err, tc.want)
        }
    })
}
```

Use field names in struct literals. Omit zero-value fields irrelevant to the test case.

### t.Error vs t.Fatal

- `t.Error` — keeps test running (prefer this).
- `t.Fatal` — only when subsequent failures would be meaningless.
- **Never call `t.Fatal` from a goroutine** — use `t.Error` and return.

### Test Helpers

- Mark with `t.Helper()`.
- Test helpers call `t.Fatal` on setup failure (environment broke, not code).
- Keep trivial assertion helpers unexported and local to the test file.

### Comparing Results

- Use `cmp.Diff` for structs, not field-by-field comparison.
- Compare stable semantic data, not formatted/serialized output.

---

## Concurrency

### Goroutine Lifetimes

- Make exit conditions obvious.
- Use `context.Context` for cancellation.
- Use `sync.WaitGroup` to ensure completion.
- Never start goroutines without a clear exit mechanism.

### Context

- Always first parameter: `func Foo(ctx context.Context, ...)`.
- Never store in a struct — pass to each method.
- Use `context.Background()` only in `main`, `init`, and test functions.

### Copying

- Never copy `sync.Mutex` or types with pointer-receiver methods.
- Be cautious with `bytes.Buffer` copies (aliased slice).

---

## Formatting

- All code must pass `gofmt`.
- No fixed line length — refactor over splitting.
- Don't split lines before indentation changes or to fit long URLs.
- Compare variable on left: `if result == "foo"` not `if "foo" == result`.
- No redundant `break` in `switch` cases (Go doesn't fall through by default).

---

## Package Design

- No "one type per file" convention — group related code logically.
- Avoid monolithic single-package designs; separate conceptually distinct code.
- Use `doc.go` for lengthy package documentation.