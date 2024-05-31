package filter

import (
	"github.com/blugelabs/bluge/analysis"
	"github.com/mozillazg/go-unidecode"
)

type NormalizeUnidecodeFilter struct{}

func NewNormalizeUnidecodeFilter() *NormalizeUnidecodeFilter {
	return &NormalizeUnidecodeFilter{}
}

func (f *NormalizeUnidecodeFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	for i, token := range input {
		decoded := unidecode.Unidecode(string(token.Term))
		token.Term = []byte(decoded)
		input[i] = token
	}
	return input
}
