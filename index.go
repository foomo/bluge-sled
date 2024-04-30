package sled

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type Index struct {
	ic     IndexConfig
	shards map[int]*shard
}

func NewIndex(ic IndexConfig) (*Index, error) {
	var err error
	shards := make(map[int]*shard, ic.ShardNum)
	for i := range ic.ShardNum {
		shards[i], err = newShard(i, ic)
		if err != nil {
			return nil, err
		}
	}
	return &Index{ic, shards}, nil
}

func (i Index) BulkInsert(data []map[string]any) error {
	start := time.Now()
	defer func() {
		slog.Debug("bulk insert complete", "duration", time.Since(start))
	}()
	// bluge does not handle document uniqueness
	uniqueData := lo.UniqBy(data, func(datum map[string]any) string {
		return fmt.Sprint(datum[i.ic.IdField])
	})
	dataByShardId := make(map[int][]map[string]any, i.ic.ShardNum)
	for _, datum := range uniqueData {
		shardId := getShardId(i.ic.ShardNum, fmt.Sprint(datum[i.ic.IdField]))
		dataByShardId[shardId] = append(dataByShardId[shardId], datum)
	}
	eg := errgroup.Group{}
	for shardId, data := range dataByShardId {
		eg.Go(func() error {
			return i.shards[shardId].BatchInsert(data)
		})
	}
	return eg.Wait()
}

// do not use
func (i Index) search_(ctx context.Context, query string, sc SearchConfig) (SearchResult, error) {
	start := time.Now()
	var combined SearchResult
	eg := errgroup.Group{}
	workers := 20
	sc.Limit = 1000
	resultChan := make(chan SearchResult, workers)
	for wi := range workers {
		from := wi * sc.Limit
		eg.Go(func() error {
			sc.From = from
			// do shard searches
			sr, err := i.shards[0].Search(ctx, query, sc)
			if err != nil {
				return err
			}
			resultChan <- sr
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return combined, err
	}
	close(resultChan)
	// combine results
	for sr := range resultChan {
		// slog.Debug(query, "hits", len(sr.Hits))
		combined.Hits = append(combined.Hits, sr.Hits...)
		combined.HitNumber += sr.HitNumber
	}
	// sort combined hits by score
	slices.SortFunc(combined.Hits, func(a Hit, b Hit) int {
		if a.Score < b.Score {
			return 1
		}
		if a.Score > b.Score {
			return -1
		}
		return 0
	})
	if len(combined.Hits) > 0 {
		combined.MaxScore = combined.Hits[0].Score
	}
	combined.Query = query
	combined.Duration = time.Since(start)
	slog.Debug("index", "query", query, "hits", combined.HitNumber, "max-score", combined.MaxScore, "duration", combined.Duration)
	return combined, nil
}

func (i Index) Search(ctx context.Context, query string, sc SearchConfig) (SearchResult, error) {
	start := time.Now()
	var combined SearchResult
	resultChan := make(chan SearchResult, i.ic.ShardNum)
	eg := errgroup.Group{}
	for _, shard := range i.shards {
		eg.Go(func() error {
			// do shard searches
			sr, err := shard.Search(ctx, query, sc)
			if err != nil {
				return err
			}
			resultChan <- sr
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return combined, err
	}
	close(resultChan)
	// combine results
	for sr := range resultChan {
		// slog.Debug(query, "hits", len(sr.Hits))
		combined.Hits = append(combined.Hits, sr.Hits...)
		combined.HitNumber += sr.HitNumber
	}
	// sort combined hits by score
	slices.SortFunc(combined.Hits, func(a Hit, b Hit) int {
		if a.Score < b.Score {
			return 1
		}
		if a.Score > b.Score {
			return -1
		}
		return 0
	})
	if len(combined.Hits) > 0 {
		combined.MaxScore = combined.Hits[0].Score
	}
	combined.Query = query
	combined.Duration = time.Since(start)
	slog.Debug(query, "hits", combined.HitNumber, "max-score", combined.MaxScore, "duration", combined.Duration)
	return combined, nil
}

type Hit struct {
	Id     string
	Score  float64
	Values map[string]string
}

type SearchResult struct {
	HitNumber uint64
	MaxScore  float64
	Duration  time.Duration
	Query     string
	Hits      []Hit
}
