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

func (i Index) BatchInsert(data []map[string]any) error {
	start := time.Now()
	defer func() {
		slog.Debug("batch insert complete", "len", len(data), "duration", time.Since(start))
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

func (i Index) Update(datum map[string]any) error {
	id := fmt.Sprint(datum[i.ic.IdField])
	slog.Debug("update", "id", id)
	shardId := getShardId(i.ic.ShardNum, id)
	return i.shards[shardId].Update(id, datum)
}

func (i Index) Upsert(data []map[string]any) error {
	// note: not truly an upsert
	slog.Debug("upsert", "len", len(data))
	var nonExisting []map[string]any
	for _, datum := range data {
		if err := i.Update(datum); err != nil {
			// TODO check for specific "does not exist" error message
			nonExisting = append(nonExisting, datum)
		}
	}
	return i.BatchInsert(nonExisting)
}

func (i Index) BatchDelete(ids []string) error {
	start := time.Now()
	defer func() {
		slog.Debug("batch delete complete", "len", len(ids), "duration", time.Since(start))
	}()
	idsByShardId := make(map[int][]string, i.ic.ShardNum)
	for _, id := range ids {
		shardId := getShardId(i.ic.ShardNum, id)
		idsByShardId[shardId] = append(idsByShardId[shardId], id)
	}
	eg := errgroup.Group{}
	for shardId, ids := range idsByShardId {
		eg.Go(func() error {
			return i.shards[shardId].BatchDelete(ids)
		})
	}
	return eg.Wait()
}

// purge any saved paths
func (i *Index) Purge() error {
	for id, shard := range i.shards {
		if err := shard.Purge(); err != nil {
			slog.Warn("failed purging shard", "id", id, "error", err)
		}
	}
	i.shards = nil
	return nil
}

func (i Index) Search(ctx context.Context, query string, sc *SearchConfig) (combined SearchResult, err error) {
	if sc == nil {
		return combined, fmt.Errorf("you must provide a valid SearchConfig")
	}
	start := time.Now()
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
