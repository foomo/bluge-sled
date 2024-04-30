package filter

import (
	"slices"
	"strings"

	"github.com/blugelabs/bluge/analysis"
)

type SynonymFilter struct {
	mapping map[string][]string
}

// provide a map of synonyms to a single word
// eg: "mixer": []string{"hochleistungsmixer", "handmixer", "stabmixer"}
func NewSynonymFilter(mapping map[string][]string) *SynonymFilter {
	return &SynonymFilter{mapping}
}

func (f *SynonymFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	for _, token := range input {
		for key, synonyms := range f.mapping {
			if slices.Contains(synonyms, string(token.Term)) {
				terms := strings.Split(key, " ")
				for i, term := range terms {
					if i == 0 {
						token.Term = []byte(term)
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
