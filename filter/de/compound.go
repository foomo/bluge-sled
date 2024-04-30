package de

import (
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/token"
)

func NewCompoundFilter() *token.DictionaryCompoundFilter {
	tm := analysis.NewTokenMap()
	for _, word := range defaultDict {
		tm.AddToken(word)
	}
	return token.NewDictionaryCompoundFilter(tm, 3, 3, 15, true)
}
