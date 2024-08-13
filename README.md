

#	# index configuration
```
type IndexConfig struct {
	ShardNum    int                           // number of shards to use
	ShardPath   string                        // filepath to store shard index (if not in-memory)
	IdField     string                        // data field to be used as doc _id
	StoreFields []string                      // fields to be stored in index; if not set, just use composite "_all"
	AnalyzerConfig analyzer.ConfigMap // analyzer config to use per field. use "*" for any field
}
```

## search configuration
```
type SearchConfig struct {
	Limit          int                           // limit number of results returned; 0 will return all
	From           int                           // offset for paging results (to be used with limit)
	SearchFields   []string                      // fields to search the query; if not set, search composite "_all"
	ReturnFields   []string                      // stored fields to return when getting search results; see IndexConfig.StoreFields to manage fields youre storing
	QueryConfig    QueryConfig                   // this will have no effect if SearchConfig.SearchFields are not set
	ScoreThreshold float64                       // filter results below specified score. if not set, includes all
	AnalyzerConfig analyzer.ConfigMap            // analyzer config to use per field. use "*" for any field
}
```

## Quickstart

### use defaults
```go
// example with german analyzer config and default index and search config
ac := analyzer.NewAnalyzerConfig(analyzer.German).WithoutStem().WithLength(3, 15)
// make sure to use proper field names in index and search config
indexConfig := sled.NewDefaultIndexConfig("my-index", "id", false, *ac)
// in this case were using all of the fields to search and returning only "image", "title", "infos", "brand"
searchConfig := sled.NewDefaultSearchConfig(*ac, nil, []string{"image", "title", "infos", "brand"})
```

### initialize the index
```go
index, err := sled.NewIndex(indexConfig)
```
### load data into the index
```go
f, err := os.Open("data.json")
if err != nil {
  return nil, err
}
defer f.Close()
var data []map[string]interface{}
if err := json.NewDecoder(f).Decode(&data); err != nil {
  return nil, err
}
if err := index.BulkInsert(data); err != nil {
  return nil, err
}
```
### search the index
```go
results, err := index.Search(ctx, q, searchConfig)
for _, hit := range results.Hits {
  // do something with the hits
}
```

### Language support note
 - In the current implementation its advised to use a single language per index
 - For multiple languages in a single index, for valid results, one would need to specify language dependant fields (and thus cannot use _all for searching)