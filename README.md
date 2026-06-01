# wildcard

Minimal wildcard matcher for Go.

Supports only `*` as the wildcard. Designed for deterministic matching and reverse lookup with predictable priority.

---

## Rules

Only `*` has special meaning. All other characters are literals.

Patterns are evaluated as:

- split by `*` into ordered parts
- parts must appear in order in the input string
- no leading `*` → prefix anchored
- no trailing `*` → suffix anchored

Examples:

| Pattern | Meaning |
|---|---|
| `abc` | exact match |
| `abc*` | must start with `abc` |
| `*abc` | must end with `abc` |
| `ab*cd` | must start with `ab` and end with `cd` |
| `a*b*c` | `a`, `b`, `c` appear in order |
| `*` | matches everything |

---

## Priority

Each pattern is ranked by:

- `Fixed` = number of matched non-`*` characters
- `Cards` = number of `*` wildcard markers in the pattern
- `Index` = lower original pattern index wins when `Fixed` and `Cards` tie

Sorting rules:

```text
1. higher Fixed
2. fewer Cards
3. lower Index
```

Example:

```text
input: abcde

abcde   Fixed = 5, Cards = 0, Index = 5
ab*de   Fixed = 4, Cards = 1, Index = 4
ab*e    Fixed = 3, Cards = 1, Index = 3
a*c*e   Fixed = 3, Cards = 2, Index = 2
a*e     Fixed = 2, Cards = 1, Index = 1
*       Fixed = 0, Cards = 1, Index = 0
```

Priority:

```text
abcde > ab*de > ab*e > a*c*e > a*e > *
```

---

## API

```go
type Match struct {
    Pattern string

    Fixed int
    Cards int
    Index int
}
```

## Example

```go
package main

import (
    "fmt"

    "github.com/chg1f/wildcard"
)

func main() {
    finder, err := wildcard.NewFinder([]string{
        "*",
        "a*e",
        "a*c*e",
        "ab*e",
        "ab*de",
        "abcde",
    })
    if err != nil {
        panic(err)
    }

    match := finder.Find("abcde")
    if match == nil {
        fmt.Println("no match")
        return
    }

    fmt.Printf(
        "pattern=%q fixed=%d cards=%d index=%d\n",
        match.Pattern,
        match.Fixed,
        match.Cards,
        match.Index,
    )

    // Output:
    // pattern="abcde" fixed=5 cards=0 index=5
}
```

```go
finder, err := wildcard.NewFinder([]string{"*", "ab*de", "abcde"})
if err != nil {
    // handle error
}

match := finder.Find("abcde")
if match != nil {
    // match.Index == 2
    // match.Pattern == "abcde"
    // match.Fixed == 5
    // match.Cards == 0
}
```

The convenience function compiles patterns and returns the best match:

```go
match, err := wildcard.Find("abcde", []string{"*", "ab*de", "abcde"})
```

---

## Lookup

`Finder` compiles patterns once for repeated lookup.

Patterns with prefix anchors are indexed by prefix, and patterns with suffix anchors are indexed by reversed suffix. Patterns that cannot be indexed, such as `*` or `*a*`, are checked as fallback candidates.

---

## Benchmark

Run:

```bash
go test -bench=BenchmarkFind -benchmem -count=5
```

Benchmark data:

- `10000` patterns
- `10000` input strings
- input string length: `100`
- pattern length: about `50`
- pattern pool includes `*suffix`, `prefix*`, `*`, and randomized multi-star patterns such as `ab*cd*ef`
- random seed is fixed for reproducible benchmark data

Result on this machine:

```text
goos: darwin
goarch: arm64
pkg: github.com/chg1f/wildcard
cpu: Apple M5
BenchmarkFind-10    592665    1832 ns/op    10304 B/op    2 allocs/op
BenchmarkFind-10    643918    1878 ns/op    10304 B/op    2 allocs/op
BenchmarkFind-10    557007    1817 ns/op    10304 B/op    2 allocs/op
BenchmarkFind-10    635924    1820 ns/op    10304 B/op    2 allocs/op
BenchmarkFind-10    601900    1849 ns/op    10304 B/op    2 allocs/op
```
