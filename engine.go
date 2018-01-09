package hypermatcher

// Engine is the hyperscanner pattern matching interface
type Engine interface {
	// Update rebuilds the pattern database, returning an optional error
	Update(patterns []string) error
	// Match takes a vectored byte corpus and returns a slice of patterns that matched the corpus and an optional error
	Match(corpus [][]byte) ([]string, error)
	// MatchStrings takes a vectored string corpus and returns a slice of patterns that matched the corpus and an optional error
	MatchStrings(corpus []string) ([]string, error)
}
