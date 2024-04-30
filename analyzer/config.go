package analyzer

import (
	"github.com/blugelabs/bluge/analysis"
)

type ConfigMap map[string]Config

func (cm ConfigMap) GetAnalyzer(key string) *analysis.Analyzer {
	if c, ok := cm[key]; ok {
		return c.GetAnalyzer()
	}
	return nil
}

func (cm ConfigMap) GetAnalyzers() map[string]*analysis.Analyzer {
	as := make(map[string]*analysis.Analyzer)
	for key, c := range cm {
		as[key] = c.GetAnalyzer()
	}
	return as
}
