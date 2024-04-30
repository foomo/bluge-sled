package sled

import (
	"strings"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
)

func newMultiFieldQuery(query string, fields []string, qc QueryConfig, as map[string]*analysis.Analyzer) bluge.Query {
	if strings.TrimSpace(query) == "" {
		return bluge.NewMatchAllQuery()
	}
	bq := bluge.NewBooleanQuery()
	for _, field := range fields {
		bq.AddShould(getQuery(query, field, qc, as))
	}
	return bq
}

func newAllFieldsQuery(query string, qc QueryConfig, as map[string]*analysis.Analyzer) bluge.Query {
	if strings.TrimSpace(query) == "" {
		return bluge.NewMatchAllQuery()
	}
	return getQuery(query, "_all", qc, as)
}

func getQuery(query, field string, qc QueryConfig, as map[string]*analysis.Analyzer) bluge.Query {
	a, ok := as[field]
	if !ok {
		a = as["*"]
	}
	q := bluge.NewMatchQuery(query).
		SetAnalyzer(a).
		SetField(field)
	if qc.GetFuzzyness(field) != 1 {
		q.SetFuzziness(2)
	}
	if qc.GetBoost(field) != 0 {
		q.SetBoost(qc.GetBoost(field))
	}
	return q
}

type QueryConfig struct {
	ImproveFuzziness map[string]bool    // improve fuzziness when searching specific fields
	FieldBoost       map[string]float64 // boost results when searching specific fields
}

func (qc QueryConfig) GetBoost(field string) float64 {
	fb, ok := qc.FieldBoost[field]
	if !ok || field == "" {
		return 0
	}
	return fb
}

func (qc QueryConfig) GetFuzzyness(field string) int {
	if ff, ok := qc.ImproveFuzziness[field]; ok && ff {
		return 2
	}
	if ff, ok := qc.ImproveFuzziness["*"]; ok && ff {
		return 2
	}
	return 1
}
