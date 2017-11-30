package hypermatcher

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/flier/gohs/hyperscan"
)

var (
	// ErrNotLoaded is returned when Match() is invoked while the pattern database is not compiled and loaded
	ErrNotLoaded = errors.New("database not loaded")
	// ErrNoPatterns is returned when Update() is invoked with an empty pattern slice
	ErrNoPatterns = errors.New("no patterns specified")
)

func compilePatterns(patterns []string) ([]*hyperscan.Pattern, error) {
	// compile patterns and add them to the internal list, returning
	// an error on the first pattern that fails to parse
	var compiledPatterns = make([]*hyperscan.Pattern, len(patterns))
	for idx, pattern := range patterns {
		var compiledPattern, compileErr = hyperscan.ParsePattern(pattern)
		if compileErr != nil {
			return nil, fmt.Errorf("error parsing pattern %s: %s", pattern, compileErr.Error())
		}

		compiledPattern.Id = idx
		compiledPatterns[idx] = compiledPattern
	}

	return compiledPatterns, nil
}

func buildDatabase(patterns []*hyperscan.Pattern) (hyperscan.VectoredDatabase, error) {
	// initialize a new database with the new patterns
	var builder = &hyperscan.DatabaseBuilder{
		Patterns: patterns,
		Mode:     hyperscan.VectoredMode,
		Platform: hyperscan.PopulatePlatform(),
	}
	var db, err = builder.Build()
	if err != nil {
		return nil, fmt.Errorf("error updating pattern database: %s", err.Error())
	}

	return db.(hyperscan.VectoredDatabase), nil
}

var matchHandler = func(id uint, from, to uint64, flags uint, context interface{}) error {
	var matched = context.(*[]uint)
	*matched = append(*matched, id)

	return nil
}

func matchedIdxToStrings(matched []uint, patterns []*hyperscan.Pattern) []string {
	var matchedSieve = make(map[uint]struct{}, 0)
	for _, patIdx := range matched {
		matchedSieve[patIdx] = struct{}{}
	}

	var matchedPatterns = make([]string, len(matchedSieve))
	var matchPatternsIdx int
	for patternsIdx := range matchedSieve {
		matchedPatterns[matchPatternsIdx] = patterns[patternsIdx].Expression.String()
		matchPatternsIdx++
	}

	return matchedPatterns
}

func stringsToBytes(corpus []string) [][]byte {
	var corpusBlocks = make([][]byte, len(corpus))
	for idx, corpusElement := range corpus {
		corpusBlocks[idx] = stringToByteSlice(corpusElement)
	}

	return corpusBlocks
}

// naughty zero copy string to []byte conversion
func stringToByteSlice(input string) []byte {
	var stringHeader = (*reflect.StringHeader)(unsafe.Pointer(&input))
	var sliceHeader = reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}

	return *(*[]byte)(unsafe.Pointer(&sliceHeader))
}
