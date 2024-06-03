package analyzer

import (
	"unicode"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/lang/de"
	"github.com/blugelabs/bluge/analysis/lang/en"
	"github.com/blugelabs/bluge/analysis/lang/fr"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
	"github.com/foomo/bluge-sled/filter"
	filterde "github.com/foomo/bluge-sled/filter/de"
)

type Tokenizer string

const (
	DigitTokenizer        Tokenizer = "digit"
	LetterTokenizer       Tokenizer = "letter"
	AlphaNumericTokenizer Tokenizer = "alpha_numeric"
	WhitespaceTokenizer   Tokenizer = "whitespace"
)

type TokenFilter string

const (
	LowercaseFilter TokenFilter = "lowercase"
	NormalizeFilter TokenFilter = "normalize"
	CompoundFilter  TokenFilter = "compound"
	StemFilter      TokenFilter = "stem"
	StopWordFilter  TokenFilter = "stop_word"
	UniqueFilter    TokenFilter = "unique"
	LengthFilter    TokenFilter = "length"
	SynonymFilter   TokenFilter = "synonym"
)

type Language string

const (
	English Language = "en"
	German  Language = "de"
	French  Language = "fr"
)

type Options struct {
	CompoundFilterDictionary []string   `json:"compound_filter_dictionary,omitempty"` // set of words for the custom compound token filter
	LengthFilterMin          int        `json:"length_filter_min,omitempty"`          // minumum length for the length token filter
	LengthFilterMax          int        `json:"length_filter_max,omitempty"`          // maximum length for the length token filter
	SynonymFilterMapping     [][]string `json:"synonym_filter_mapping,omitempty"`     // word/sentence mapping for synonym token filter
	Language                 Language   `json:"language,omitempty"`
}

type Config struct {
	Tokenizer    Tokenizer     `yaml:"tokenizer,omitempty" json:"tokenizer,omitempty"`         // tokenizer to use, see enums
	TokenFilters []TokenFilter `yaml:"token_filters,omitempty" json:"token_filters,omitempty"` // token filters to use, see enums
	Options      Options       `yaml:"options,omitempty" json:"options,omitempty"`             // additional configuration options needed by some of the tokenizers/filters
}

func (ac Config) GetAnalyzer() *analysis.Analyzer {
	a := analysis.Analyzer{}
	switch ac.Tokenizer {
	case DigitTokenizer:
		a.Tokenizer = tokenizer.NewCharacterTokenizer(unicode.IsDigit)
	case LetterTokenizer:
		a.Tokenizer = tokenizer.NewCharacterTokenizer(unicode.IsLetter)
	case AlphaNumericTokenizer:
		a.Tokenizer = tokenizer.NewCharacterTokenizer(letterOrNumber)
	case WhitespaceTokenizer:
		a.Tokenizer = tokenizer.NewWhitespaceTokenizer()
	default:
		a.Tokenizer = tokenizer.NewCharacterTokenizer(unicode.IsLetter)
	}
	for _, tf := range ac.TokenFilters {
		switch tf {
		case LowercaseFilter:
			a.TokenFilters = append(a.TokenFilters, token.NewLowerCaseFilter())
		case CompoundFilter:
			if len(ac.Options.CompoundFilterDictionary) > 0 {
				// use custom dictionary for compounder
				a.TokenFilters = append(a.TokenFilters, filter.NewCompoundFilter(
					ac.Options.CompoundFilterDictionary))
			} else {
				// otherwise use a compound filter by language
				f := newCompoundFilter(ac.Options.Language)
				if f != nil {
					a.TokenFilters = append(a.TokenFilters, f)
				}
			}
		case NormalizeFilter:
			f := newNormalizeFilter(ac.Options.Language)
			if f != nil {
				a.TokenFilters = append(a.TokenFilters, f)
			}
		case StemFilter:
			f := newStemFilter(ac.Options.Language)
			if f != nil {
				a.TokenFilters = append(a.TokenFilters, f)
			}
		case StopWordFilter:
			a.TokenFilters = append(a.TokenFilters, newStopWordFilter(ac.Options.Language))
		case UniqueFilter:
			a.TokenFilters = append(a.TokenFilters, token.NewUniqueTermFilter())
		case LengthFilter:
			if ac.Options.LengthFilterMax-ac.Options.LengthFilterMin < 2 {
				continue
			}
			a.TokenFilters = append(a.TokenFilters, token.NewLengthFilter(
				ac.Options.LengthFilterMin, ac.Options.LengthFilterMax))
		case SynonymFilter:
			if len(ac.Options.SynonymFilterMapping) == 0 {
				continue
			}
			a.TokenFilters = append(a.TokenFilters, filter.NewSynonymFilter(
				ac.Options.SynonymFilterMapping))
		}
	}
	return &a
}

func newNormalizeFilter(l Language) analysis.TokenFilter {
	switch l {
	case German:
		return de.NormalizeFilter()
	default:
		return filter.NewNormalizeUnidecodeFilter()
	}
}

func newStopWordFilter(l Language) analysis.TokenFilter {
	switch l {
	case German:
		return de.StopWordsFilter()
	case French:
		return fr.StopWordsFilter()
	default:
		return en.StopWordsFilter()
	}
}

func newStemFilter(l Language) analysis.TokenFilter {
	switch l {
	case German:
		return de.StemmerFilter()
	case French:
		return fr.StemmerFilter()
	default:
		return en.StemmerFilter()
	}
}

func newCompoundFilter(l Language) analysis.TokenFilter {
	switch l {
	case German:
		return filterde.NewCompoundFilter()
	default:
		return nil
	}
}
