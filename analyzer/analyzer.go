package analyzer

import (
	"unicode"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/lang/de"
	"github.com/blugelabs/bluge/analysis/lang/en"
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
	LowercaseFilter       TokenFilter = "lowercase"
	CustomCompoundFilter  TokenFilter = "compound_custom"
	GermanCompoundFilter  TokenFilter = "compound_de"
	GermanStemFilter      TokenFilter = "stem_de"
	EnglishStemFilter     TokenFilter = "stem_en"
	GermanStopWordFilter  TokenFilter = "stop_word_de"
	EnglishStopWordFilter TokenFilter = "stop_word_en"
	UniqueFilter          TokenFilter = "unique"
	LengthFilter          TokenFilter = "length"
	SynonymFilter         TokenFilter = "synonym"
)

type Options struct {
	CompoundFilterDictionary []string            `json:"compound_filter_dictionary,omitempty"` // set of words for the custom compound token filter
	LengthFilterMin          int                 `json:"length_filter_min,omitempty"`          // minumum length for the length token filter
	LengthFilterMax          int                 `json:"length_filter_max,omitempty"`          // maximum length for the length token filter
	SynonymFilterMapping     map[string][]string `json:"synonym_filter_mapping,omitempty"`     // word/sentence mapping for synonym token filter
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
		case CustomCompoundFilter:
			if len(ac.Options.CompoundFilterDictionary) == 0 {
				continue
			}
			a.TokenFilters = append(a.TokenFilters, filter.NewCompoundFilter(
				ac.Options.CompoundFilterDictionary))
		case GermanCompoundFilter:
			a.TokenFilters = append(a.TokenFilters, filterde.NewCompoundFilter())
		case GermanStemFilter:
			a.TokenFilters = append(a.TokenFilters, de.LightStemmerFilter())
		case EnglishStemFilter:
			a.TokenFilters = append(a.TokenFilters, en.StemmerFilter())
		case GermanStopWordFilter:
			a.TokenFilters = append(a.TokenFilters, de.StopWordsFilter())
		case EnglishStopWordFilter:
			a.TokenFilters = append(a.TokenFilters, en.StopWordsFilter())
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
