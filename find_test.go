package wildcard

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var benchmarkMatch *Match

func TestFind(t *testing.T) {
	cases := []struct {
		Name string

		String   string
		Patterns []string

		Expected *Match
	}{
		{
			Name:     "best priority",
			String:   "abcde",
			Patterns: []string{"*", "a*e", "a*c*e", "ab*e", "ab*de", "abcde"},
			Expected: &Match{
				Index:   5,
				Pattern: "abcde",
				Fixed:   5,
				Cards:   0,
			},
		},
		{
			Name:     "exact",
			String:   "abc",
			Patterns: []string{"ab", "abc", "*abc"},
			Expected: &Match{
				Index:   1,
				Pattern: "abc",
				Fixed:   3,
				Cards:   0,
			},
		},
		{
			Name:     "prefix anchored",
			String:   "abcdef",
			Patterns: []string{"bc*", "abc*"},
			Expected: &Match{
				Index:   1,
				Pattern: "abc*",
				Fixed:   3,
				Cards:   1,
			},
		},
		{
			Name:     "suffix anchored",
			String:   "abcdef",
			Patterns: []string{"*cde", "*def"},
			Expected: &Match{
				Index:   1,
				Pattern: "*def",
				Fixed:   3,
				Cards:   1,
			},
		},
		{
			Name:     "ordered parts",
			String:   "abcdef",
			Patterns: []string{"a*d*f", "a*f*d"},
			Expected: &Match{
				Index:   0,
				Pattern: "a*d*f",
				Fixed:   3,
				Cards:   2,
			},
		},
		{
			Name:     "no match",
			String:   "abcdef",
			Patterns: []string{"bc", "ab*d"},
		},
		{
			Name:     "tie keeps earlier pattern",
			String:   "abc",
			Patterns: []string{"a*c", "a*b"},
			Expected: &Match{
				Index:   0,
				Pattern: "a*c",
				Fixed:   2,
				Cards:   1,
			},
		},
		{
			Name:     "details",
			String:   "abcde",
			Patterns: []string{"*", "ab*de"},
			Expected: &Match{
				Index:   1,
				Pattern: "ab*de",
				Fixed:   4,
				Cards:   1,
			},
		},
		{
			Name:     "control characters",
			String:   "abcd\x00efg",
			Patterns: []string{"abcd*efg", "abcd\x00efg", "abcd\x00*", "*"},
			Expected: &Match{
				Index:   1,
				Pattern: "abcd\x00efg",
				Fixed:   8,
				Cards:   0,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			finder, err := NewFinder(c.Patterns)
			require.NoError(t, err)

			actual := finder.Find(c.String)
			require.Equal(t, c.Expected, actual)
		})
	}
}

func BenchmarkFind(b *testing.B) {
	const patternsTotal = 10000
	const stringsTotal = 10000
	const inputLength = 100
	const fixedLength = 49

	rng := rand.New(rand.NewSource(1))
	inputs := make([]string, 0, stringsTotal)
	for i := 0; i < stringsTotal; i++ {
		inputs = append(inputs, fmt.Sprintf("svc-%05d-%s", i, strings.Repeat(string(rune('a'+i%26)), inputLength-10)))
	}

	patterns := make([]string, 0, patternsTotal)
	for i := 0; i < patternsTotal-1; i++ {
		input := inputs[i%len(inputs)]
		switch i % 3 {
		case 0:
			patterns = append(patterns, "*"+input[len(input)-fixedLength:])
		case 1:
			patterns = append(patterns, input[:fixedLength]+"*")
		default:
			pattern := []byte(input[:fixedLength+1])
			stars := 2 + rng.Intn(4)
			for n := 0; n < stars; {
				index := rng.Intn(len(pattern))
				if pattern[index] == '*' {
					continue
				}
				pattern[index] = '*'
				n++
			}
			patterns = append(patterns, string(pattern))
		}
	}
	patterns = append(patterns, "*")

	finder, err := NewFinder(patterns)
	require.NoError(b, err)
	require.NotNil(b, finder.Find(inputs[0]))

	b.ReportAllocs()
	b.ResetTimer()

	var match *Match
	for i := 0; i < b.N; i++ {
		match = finder.Find(inputs[i%len(inputs)])
	}
	benchmarkMatch = match
}
