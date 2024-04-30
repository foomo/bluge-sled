package filter

import (
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/token"
)

func NewCompoundFilter(dict []string) *token.DictionaryCompoundFilter {
	tm := analysis.NewTokenMap()
	for _, word := range dict {
		tm.AddToken(word)
	}
	return token.NewDictionaryCompoundFilter(tm, 3, 3, 15, true)
}
