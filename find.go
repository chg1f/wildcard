package wildcard

import (
	"fmt"
	"sort"
	"strings"
)

const Wildcard = "*"

// Match describes a matched pattern and its priority data.
type Match struct {
	// Pattern is the matched wildcard pattern.
	Pattern string

	// Fixed is the number of matched non-wildcard characters.
	Fixed int
	// Cards is the number of wildcard markers in the pattern.
	Cards int
	// Index is the original index of the matched pattern.
	Index int
}

// compile stores precomputed data for one pattern.
type compile struct {
	// index is the pattern's original position in Finder.patterns.
	index int
	// parts contains the pattern split by wildcard markers.
	parts []string
	// length is the total number of non-wildcard characters.
	length int
	// cards is the number of wildcard markers.
	cards int
	// prefix reports whether the pattern starts with a wildcard marker.
	prefix bool
	// suffix reports whether the pattern ends with a wildcard marker.
	suffix bool
}

func (p compile) indexed() bool {
	return (!p.prefix && p.parts[0] != "") ||
		(!p.suffix && p.parts[len(p.parts)-1] != "")
}

func (p compile) match(s string) bool {
	if p.cards == 0 {
		return len(p.parts) == 1 && s == p.parts[0]
	}

	pos := 0
	start := 0
	end := len(p.parts)

	if !p.prefix {
		prefix := p.parts[0]
		if !strings.HasPrefix(s, prefix) {
			return false
		}
		pos = len(prefix)
		start = 1
	}

	if !p.suffix {
		end--
	}

	for i := start; i < end; i++ {
		part := p.parts[i]
		if part == "" {
			continue
		}

		index := strings.Index(s[pos:], part)
		if index == -1 {
			return false
		}
		pos += index + len(part)
	}

	if p.suffix {
		return true
	}

	suffix := p.parts[len(p.parts)-1]
	if !strings.HasSuffix(s, suffix) {
		return false
	}

	return pos <= len(s)-len(suffix)
}

// node is a trie node used by prefix and reversed suffix indexes.
type node struct {
	// patterns contains original pattern indexes that end at this node.
	patterns []int
	// children maps the next byte to a child trie node.
	children map[byte]*node
}

func (n *node) add(key string, patternIndex int) {
	for i := 0; i < len(key); i++ {
		if n.children == nil {
			n.children = make(map[byte]*node)
		}

		next := n.children[key[i]]
		if next == nil {
			next = &node{}
			n.children[key[i]] = next
		}
		n = next
	}

	n.patterns = append(n.patterns, patternIndex)
}

func (n *node) addReverse(key string, patternIndex int) {
	for i := len(key) - 1; i >= 0; i-- {
		if n.children == nil {
			n.children = make(map[byte]*node)
		}

		next := n.children[key[i]]
		if next == nil {
			next = &node{}
			n.children[key[i]] = next
		}
		n = next
	}

	n.patterns = append(n.patterns, patternIndex)
}

func (n *node) collectPrefixes(s string, candidates []bool) {
	for i := 0; i < len(s); i++ {
		next := n.children[s[i]]
		if next == nil {
			return
		}

		n = next
		for _, patternIndex := range n.patterns {
			candidates[patternIndex] = true
		}
	}
}

func (n *node) collectSuffixes(s string, candidates []bool) {
	for i := len(s) - 1; i >= 0; i-- {
		next := n.children[s[i]]
		if next == nil {
			return
		}

		n = next
		for _, patternIndex := range n.patterns {
			candidates[patternIndex] = true
		}
	}
}

// Finder finds the highest-priority wildcard pattern for an input string.
type Finder struct {
	// patterns contains the original patterns.
	patterns []string
	// compiles contains compiled patterns sorted by priority.
	compiles []compile

	// exact maps exact patterns to their highest-priority compile.
	exact map[string]compile
	// prefix indexes anchored pattern prefixes.
	prefix *node
	// suffix indexes anchored pattern suffixes in reverse byte order.
	suffix *node
	// unindexed contains original indexes for patterns that cannot use trie indexes.
	unindexed []int
}

// NewFinder compiles patterns for repeated lookups.
func NewFinder(patterns []string) (*Finder, error) {
	f := &Finder{
		patterns: append([]string(nil), patterns...),
		compiles: make([]compile, 0, len(patterns)),
		exact:    make(map[string]compile),
		prefix:   &node{},
		suffix:   &node{},
	}

	unique := make(map[string]struct{}, len(patterns))
	for index := range patterns {
		if _, ok := unique[patterns[index]]; ok {
			return nil, fmt.Errorf("%s", patterns[index])
		}
		unique[patterns[index]] = struct{}{}

		cards := strings.Count(patterns[index], Wildcard)
		parts := strings.Split(patterns[index], Wildcard)
		length := 0
		for _, part := range parts {
			length += len(part)
		}
		compiled := compile{
			index:  index,
			parts:  parts,
			length: length,
			cards:  cards,
			prefix: strings.HasPrefix(patterns[index], Wildcard),
			suffix: strings.HasSuffix(patterns[index], Wildcard),
		}
		f.compiles = append(f.compiles, compiled)

		if cards == 0 {
			if _, ok := f.exact[patterns[index]]; !ok {
				f.exact[patterns[index]] = compiled
			}
		}
		if !compiled.prefix && parts[0] != "" {
			f.prefix.add(parts[0], index)
		}
		if !compiled.suffix && parts[len(parts)-1] != "" {
			f.suffix.addReverse(parts[len(parts)-1], index)
		}
		if !compiled.indexed() {
			f.unindexed = append(f.unindexed, index)
		}
	}

	sort.SliceStable(f.compiles, func(i, j int) bool {
		if f.compiles[i].length != f.compiles[j].length {
			return f.compiles[i].length > f.compiles[j].length
		}
		if f.compiles[i].cards != f.compiles[j].cards {
			return f.compiles[i].cards < f.compiles[j].cards
		}
		return f.compiles[i].index < f.compiles[j].index
	})

	return f, nil
}

func (f *Finder) candidates(s string) []bool {
	candidates := make([]bool, len(f.compiles))
	for _, patternIndex := range f.unindexed {
		candidates[patternIndex] = true
	}
	f.prefix.collectPrefixes(s, candidates)
	f.suffix.collectSuffixes(s, candidates)

	return candidates
}

// Find returns the best matching pattern, or nil.
func (f *Finder) Find(s string) *Match {
	if pattern, ok := f.exact[s]; ok {
		return &Match{
			Pattern: f.patterns[pattern.index],
			Fixed:   pattern.length,
			Cards:   pattern.cards,
			Index:   pattern.index,
		}
	}

	candidates := f.candidates(s)
	for _, pattern := range f.compiles {
		if !candidates[pattern.index] {
			continue
		}

		if pattern.match(s) {
			return &Match{
				Pattern: f.patterns[pattern.index],
				Fixed:   pattern.length,
				Cards:   pattern.cards,
				Index:   pattern.index,
			}
		}
	}

	return nil
}

// Find compiles patterns and returns the best match.
func Find(s string, patterns []string) (*Match, error) {
	m, err := NewFinder(patterns)
	if err != nil {
		return nil, err
	}

	return m.Find(s), nil
}
