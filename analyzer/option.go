package analyzer

import (
	"unicode"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/lang/de"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
	"github.com/foomo/bluge-sled/filter"
	filterde "github.com/foomo/bluge-sled/filter/de"
)

func NewAnalyzerWithOptions(opts ...Option) (*analysis.Analyzer, error) {
	a := analysis.Analyzer{}
	for _, opt := range opts {
		if err := opt(&a); err != nil {
			return nil, err
		}
	}
	return &a, nil
}

type Option func(a *analysis.Analyzer) error

func UseDigitTokenizer() Option {
	return func(a *analysis.Analyzer) error {
		a.Tokenizer = tokenizer.NewCharacterTokenizer(unicode.IsDigit)
		return nil
	}
}

func UseLetterTokenizer() Option {
	return func(a *analysis.Analyzer) error {
		a.Tokenizer = tokenizer.NewCharacterTokenizer(unicode.IsDigit)
		return nil
	}
}

func UseAlphaNumericTokenizer() Option {
	return func(a *analysis.Analyzer) error {
		a.Tokenizer = tokenizer.NewCharacterTokenizer(letterOrNumber)
		return nil
	}
}

func UseLowercaseFilter() Option {
	return func(a *analysis.Analyzer) error {
		a.TokenFilters = append(a.TokenFilters, token.NewLowerCaseFilter())
		return nil
	}
}

func UseGermanCompoundFilter() Option {
	return func(a *analysis.Analyzer) error {
		a.TokenFilters = append(a.TokenFilters, filterde.NewCompoundFilter())
		return nil
	}
}

func UseCompoundFilter(dict []string) Option {
	return func(a *analysis.Analyzer) error {
		a.TokenFilters = append(a.TokenFilters, filter.NewCompoundFilter(dict))
		return nil
	}
}

func UseGermanStemFilter() Option {
	return func(a *analysis.Analyzer) error {
		a.TokenFilters = append(a.TokenFilters, de.LightStemmerFilter())
		return nil
	}
}

func UseGermanStopWordFilter() Option {
	return func(a *analysis.Analyzer) error {
		a.TokenFilters = append(a.TokenFilters, de.StopWordsFilter())
		return nil
	}
}

func UseUniqueFilter() Option {
	return func(a *analysis.Analyzer) error {
		a.TokenFilters = append(a.TokenFilters, token.NewUniqueTermFilter())
		return nil
	}
}

func UseLengthFilter(min, max int) Option {
	return func(a *analysis.Analyzer) error {
		a.TokenFilters = append(a.TokenFilters, token.NewLengthFilter(min, max))
		return nil
	}
}

func UseSynonymFilter(items [][]string) Option {
	return func(a *analysis.Analyzer) error {
		a.TokenFilters = append(a.TokenFilters, filter.NewSynonymFilter(items))
		return nil
	}
}

func letterOrNumber(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}
