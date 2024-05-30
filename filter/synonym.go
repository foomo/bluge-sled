package filter

import (
	"slices"

	"github.com/blugelabs/bluge/analysis"
)

type SynonymFilter struct {
	items [][]string
}

// provide a list of string slices as synonyms
// eg: []string{"mixer", "hochleistungsmixer", "handmixer", "stabmixer"}
func NewSynonymFilter(items [][]string) *SynonymFilter {
	return &SynonymFilter{items}
}

func (f *SynonymFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	for _, token := range input {
		for _, synonyms := range f.items {
			if slices.Contains(synonyms, string(token.Term)) {
				for _, term := range synonyms {
					if term == string(token.Term) {
						continue
					}
					newtoken := &analysis.Token{
						Term:         []byte(term),
						PositionIncr: 1,
						Start:        len(token.Term),
						End:          len(token.Term) + len(term),
						Type:         token.Type,
						KeyWord:      token.KeyWord,
					}
					input = append(input, newtoken)
				}
			}
		}
	}
	return input
}
