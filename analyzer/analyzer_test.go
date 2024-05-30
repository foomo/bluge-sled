package analyzer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/lang/de"
	"github.com/blugelabs/bluge/analysis/token"
	"github.com/blugelabs/bluge/analysis/tokenizer"
	"github.com/foomo/bluge-sled/filter"
	filterde "github.com/foomo/bluge-sled/filter/de"
)

func TestAnalyzer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// {"debug", "grÜn kochmesser", "grun kochmess koch mess ess"},
		// {"debug", "kocher rot", "grun kochmess koch mess ess"},
		// {"debug", "Rüstmesser COLORI®", "grun kochmess koch mess ess"},
		// {"debug", "Kinderbesteck", "grun kochmess koch mess ess"},
		// {"debug", "schweizermesser", "grun kochmess koch mess ess"},
		// {"debug", "besteck set silber", ""},
		{"debug", "handmixer", ""},
		// {"debug", "messer", ""},
	}
	a := &analysis.Analyzer{
		Tokenizer: tokenizer.NewCharacterTokenizer(unicode.IsLetter),
		TokenFilters: []analysis.TokenFilter{
			token.NewLowerCaseFilter(),
			filter.NewSynonymFilter([][]string{
				{"mixer", "handmixer", "stabmixer"},
			}),
			filterde.NewCompoundFilter(),
			de.LightStemmerFilter(),
			de.NormalizeFilter(),
			de.StopWordsFilter(),
			token.NewUniqueTermFilter(),
			token.NewLengthFilter(3, 15),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.Analyze([]byte(tt.input))
			var actual []string
			for _, item := range got {
				actual = append(actual, string(item.Term))
			}
			actualString := strings.Join(actual, " ")
			if !reflect.DeepEqual(actualString, tt.want) {
				t.Errorf("Analyzer().Analyze = %v, want %v", actualString, tt.want)
			}
			fmt.Sprint(got)
		})
	}
}
