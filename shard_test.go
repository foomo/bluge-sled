package sled

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShards(t *testing.T) {
	results := map[int]int{}

	test := map[string]int{}

	const numIDs = 1000000
	for range numIDs {
		id := uuid.NewString()
		shard := getShardId(4, id)
		results[shard]++
		test[id] = shard
	}

	for id, shard := range test {
		assert.Equal(t, shard, getShardId(4, id))
	}

	total := 0

	for _, shardNum := range results {
		total += shardNum
	}

	assert.Equal(t, numIDs, total)
	fmt.Println(results, total)

}
