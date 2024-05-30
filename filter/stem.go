package filter

import (
	"github.com/blugelabs/bluge/analysis"
	"github.com/tebeka/snowball"
)

type SnowballStemFilter struct {
	stemmer *snowball.Stemmer
}

func NewSnowballStemmer(language string) (*SnowballStemFilter, error) {
	stemmer, err := snowball.New(language)
	if err != nil {
		return nil, err
	}
	return &SnowballStemFilter{stemmer}, nil
}

func (f *SnowballStemFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	for _, token := range input {
		stemmedTerm := f.stemmer.Stem(string(token.Term))
		newtoken := &analysis.Token{
			Term:         []byte(stemmedTerm),
			PositionIncr: 1,
			Start:        len(token.Term),
			End:          len(token.Term) + len(stemmedTerm),
			Type:         token.Type,
			KeyWord:      token.KeyWord,
		}
		input = append(input, newtoken)
	}
	return input
}
