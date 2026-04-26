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

Each pattern is scored by:

```text
score = matched_literal_length / input_length
```

Where:

- `matched_literal_length` = number of matched non-`*` characters
- `input_length` = length of the input string

Sorting rules:

```text
1. higher score
2. fewer '*'
```

Example:

```text
input: abcde

abcde   5/5 = 1.00, stars = 0
ab*de   4/5 = 0.80, stars = 1
ab*e    3/5 = 0.60, stars = 1
a*c*e   3/5 = 0.60, stars = 2
a*e     2/5 = 0.40, stars = 1
*       0/5 = 0.00, stars = 1
```

Result order:

```text
abcde > ab*de > ab*e > a*c*e > a*e > *
```
