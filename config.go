package sled

import (
	"path/filepath"

	"github.com/foomo/bluge-sled/analyzer"
)

type Config struct {
	IndexConfig  IndexConfig
	SearchConfig SearchConfig
}

type IndexConfig struct {
	ShardNum       int                `yaml:"shard_num,omitempty" json:"shard_num,omitempty"`             // number of shards to use
	ShardPath      string             `yaml:"shard_path,omitempty" json:"shard_path,omitempty"`           // filepath to store shard index (if not in-memory)
	IdField        string             `yaml:"id_field,omitempty" json:"id_field,omitempty"`               // data field to be used as doc _id
	StoreFields    []string           `yaml:"store_fields,omitempty" json:"store_fields,omitempty"`       // fields to be stored in index; if not set, just use composite "_all"
	AnalyzerConfig analyzer.ConfigMap `yaml:"analyzer_config,omitempty" json:"analyzer_config,omitempty"` // analyzer config to use per field. use "*" for any field
}

// index config with opinionated defaults
func NewDefaultIndexConfig(name, idField string, inMemory bool, ac analyzer.Config) IndexConfig {
	ic := IndexConfig{
		ShardNum: 1,
		IdField:  idField,
		AnalyzerConfig: analyzer.ConfigMap{
			"*": ac,
		},
	}
	ic.ShardPath = filepath.Join(".", "data", name, "shard")
	if inMemory {
		ic.ShardPath = ""
	}
	return ic
}

type SearchConfig struct {
	Limit                    int                `yaml:"limit,omitempty" json:"limit,omitempty"`                                             // limit number of results returned; 0 will return all
	From                     int                `yaml:"from,omitempty" json:"from,omitempty"`                                               // offset for paging results (to be used with limit)
	SearchFields             []string           `yaml:"search_fields,omitempty" json:"search_fields,omitempty"`                             // fields to search the query; if not set, search composite "_all"
	ReturnFields             []string           `yaml:"return_fields,omitempty" json:"return_fields,omitempty"`                             // stored fields to return when getting search results; see IndexConfig.StoreFields to manage fields youre storing
	QueryConfig              QueryConfig        `yaml:"query_config,omitempty" json:"query_config,omitempty"`                               // this will have no effect if SearchConfig.SearchFields are not set
	ScoreThreshold           float64            `yaml:"score_threshold,omitempty" json:"score_threshold,omitempty"`                         // filter results below specified score. if not set, includes all
	MaxScorePercentThreshold float64            `yaml:"max_score_percent_threshold,omitempty" json:"max_score_percent_threshold,omitempty"` // filter results below specified percent of max score.
	AnalyzerConfig           analyzer.ConfigMap `yaml:"analyzer_config,omitempty" json:"analyzer_config,omitempty"`                         // analyzer config to use per field. use "*" for any field
}

// search config with opinionated defaults
// note: ImproveFuzziness will impact query speed but improves results
func NewDefaultSearchConfig(ac analyzer.Config, returnFields []string) SearchConfig {
	return SearchConfig{
		Limit: 25,
		AnalyzerConfig: analyzer.ConfigMap{
			"*": ac,
		},
		QueryConfig: QueryConfig{
			ImproveFuzziness: map[string]bool{
				"*": true,
			},
		},
		ReturnFields: returnFields,
	}
}
