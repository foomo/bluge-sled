package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	sled "github.com/foomo/bluge-sled"
	"github.com/foomo/bluge-sled/example/item"
)

func main() {
	flagDataPath := flag.String("data-src", "data/fake/data.json", "data source path")
	flagGenerateFake := flag.Bool("fake", false, "generate fake data")
	flagGenerateAmount := flag.Int("data-items", 100000, "how much data to generate")
	flagReload := flag.Bool("reload", false, "reload index")
	flag.Parse()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	if *flagGenerateFake {
		if _, err := sled.GenerateFakeData(*flagGenerateAmount, *flagDataPath); err != nil {
			log.Fatal(err)
		}
	}

	ic := sled.NewDefaultIndexConfig("fake", "id", false)
	ic.StoreFields = []string{"title", "description", "brand", "color"}
	sc := sled.SearchConfig{
		Limit:          25,
		AnalyzerConfig: ic.AnalyzerConfig,
		ReturnFields:   []string{"title", "description", "brand", "color"},
	}

	bi, err := loadIndex(*flagDataPath, ic, *flagReload)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		res, err := bi.Search(r.Context(), q, sc)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		for _, hit := range res.Hits {
			component := item.Item(
				"",
				hit.Values["title"],
				hit.Values["description"],
				hit.Values["brand"],
				hit.Values["color"])
			component.Render(r.Context(), w)
		}
	})
	addr := ":8080"
	slog.Info("started", "addr", addr)
	http.ListenAndServe(addr, nil)
}

func loadIndex(dataPath string, ic sled.IndexConfig, reload bool) (*sled.Index, error) {
	if reload {
		if ic.ShardPath == "" {
			return nil, fmt.Errorf("cannot reload if IndexConfig.ShardPath is empty")
		}
		paths, err := filepath.Glob(ic.ShardPath + "*/*")
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			os.RemoveAll(path)
		}
	}
	bi, err := sled.NewIndex(ic)
	if err != nil {
		return nil, err
	}
	if reload {
		f, err := os.Open(dataPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var data []map[string]interface{}
		if err := json.NewDecoder(f).Decode(&data); err != nil {
			return nil, err
		}
		if err := bi.BulkInsert(data); err != nil {
			return nil, err
		}
	}
	return bi, nil
}
