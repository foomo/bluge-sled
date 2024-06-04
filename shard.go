package sled

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
	"golang.org/x/sync/errgroup"
)

type shard struct {
	id int
	ic IndexConfig
	c  bluge.Config
}

func newShard(id int, ic IndexConfig) (*shard, error) {
	c := bluge.InMemoryOnlyConfig()
	if ic.ShardPath != "" {
		path := getShardPath(ic.ShardPath, id)
		c = bluge.DefaultConfig(path)
	}
	if a := ic.AnalyzerConfig.GetAnalyzer("*"); a != nil {
		c.DefaultSearchAnalyzer = a
	}
	return &shard{id, ic, c}, nil
}

func getShardPath(basePath string, id int) string {
	return fmt.Sprintf("%v-%v", basePath, id)
}

func (s shard) BatchInsert(data []map[string]any) error {
	batch, fs, err := newBatchInsert(s.id, data, s.ic.IdField, s.ic.StoreFields, s.ic.AnalyzerConfig.GetAnalyzers())
	if err != nil {
		return err
	}
	w, err := bluge.OpenWriter(s.c)
	if err != nil {
		return err
	}
	defer w.Close()
	slog.Debug("data", "fields", strings.Join(fs, ","))
	return w.Batch(batch)
}

func (s shard) Update(id string, datum map[string]any) error {
	doc, _, err := newDocument(datum, s.ic.IdField, s.ic.StoreFields, s.ic.AnalyzerConfig.GetAnalyzers())
	if err != nil {
		return err
	}
	w, err := bluge.OpenWriter(s.c)
	if err != nil {
		return err
	}
	defer w.Close()
	return w.Update(doc.ID(), doc)
}

func (s shard) BatchDelete(ids []string) error {
	w, err := bluge.OpenWriter(s.c)
	if err != nil {
		return err
	}
	defer w.Close()
	b := bluge.NewBatch()
	for _, id := range ids {
		b.Delete(bluge.Identifier(id))
	}
	return w.Batch(b)
}

func (s *shard) Purge() error {
	if s.ic.ShardPath != "" {
		paths, err := filepath.Glob(s.ic.ShardPath + "*/*")
		if err != nil {
			return err
		}
		for _, path := range paths {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		}
	}
	s = nil // after purging make shard unusable
	return nil
}

func (s shard) Search(ctx context.Context, query string, sc SearchConfig) (SearchResult, error) {
	var sr SearchResult
	r, err := bluge.OpenReader(s.c)
	if err != nil {
		return sr, err
	}
	defer r.Close()

	var q bluge.Query
	if len(sc.SearchFields) > 0 {
		q = newMultiFieldQuery(query, sc.SearchFields, sc.QueryConfig, s.ic.AnalyzerConfig.GetAnalyzers())
	} else {
		q = newAllFieldsQuery(query, sc.QueryConfig, s.ic.AnalyzerConfig.GetAnalyzers())
	}
	var req bluge.SearchRequest
	req = bluge.NewAllMatches(q).WithStandardAggregations()
	if sc.Limit != 0 {
		req = bluge.NewTopNSearch(sc.Limit, q).SetFrom(sc.From).WithStandardAggregations()
	}
	dmi, err := r.Search(ctx, req)
	if err != nil {
		return sr, err
	}
	hits, err := processMatches(dmi, sc)
	if err != nil {
		return sr, err
	}
	sr.Hits = hits
	sr.Query = query
	sr.HitNumber = uint64(len(sr.Hits))
	sr.MaxScore = dmi.Aggregations().Metric("max_score")
	sr.Duration = dmi.Aggregations().Duration()
	return sr, err
}

func processMatches(dmi search.DocumentMatchIterator, sc SearchConfig) (hits []Hit, err error) {
	for {
		match, err := dmi.Next()
		if err != nil {
			return nil, err
		}
		if match == nil {
			break
		}
		if sc.ScoreThreshold > 0 && sc.ScoreThreshold > match.Score {
			// exclude results lower than configured threshold
			break
		}
		var hit Hit
		hit.Values = make(map[string]string, 1)
		if err := match.VisitStoredFields(func(field string, value []byte) bool {
			switch true {
			case field == "_id":
				hit.Id = string(value)
				hit.Score = match.Score
			case sc.ReturnFields != nil && slices.Contains(sc.ReturnFields, field):
				hit.Values[field] = string(value)
			}
			return true
		}); err != nil {
			return nil, err
		}
		hits = append(hits, hit)
	}
	return hits, nil
}

func batchProcessResults(batchNum int, dmi search.DocumentMatchIterator, returnFields []string) (SearchResult, error) {
	eg := errgroup.Group{}
	var sr SearchResult
	var matches []*search.DocumentMatch
	var iterations int
	for {
		match, err := dmi.Next()
		if err != nil {
			return sr, err
		}
		if len(matches) == batchNum || match == nil {
			iterations++
			matchSubset := matches[batchNum*(iterations-1):]
			eg.Go(func() error {
				for _, match := range matchSubset {
					var hit Hit
					hit.Values = make(map[string]string)
					if err := match.VisitStoredFields(func(field string, value []byte) bool {
						if slices.Contains(returnFields, field) || field == "_id" {
							hit.Score = match.Score
							hit.Values[field] = string(value)
						}
						return true
					}); err != nil {
						return err
					}
					sr.Hits = append(sr.Hits, hit)
				}
				return nil
			})
		}
		if match == nil {
			break
		}
		matches = append(matches, match)
	}
	if err := eg.Wait(); err != nil {
		return sr, err
	}
	return sr, nil
}
