package sled

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"strconv"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/index"
	"github.com/cespare/xxhash"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func getShardId(numShards int, id string) int {
	return int(xxhash.Sum64String(id) % uint64(numShards))
}

func newBatchInsert(shardId int, data []map[string]any, idField string, storeFields []string, as map[string]*analysis.Analyzer) (b *index.Batch, fields []string, err error) {
	b = bluge.NewBatch()
	slog.Debug("bulk inserting data", "shard", shardId, "length", len(data))
	for i, datum := range data {
		var doc *bluge.Document
		doc, fields, err = newDocument(datum, idField, storeFields, as)
		if err != nil {
			// todo warn or quit?
			return nil, nil, errors.WithMessagef(err, "failed for item at index %d", i)
		}
		b.Insert(doc)
	}
	return b, lo.Uniq(fields), nil
}

func newIndex(data []map[string]any, iw *bluge.Writer, idField string, storeFields []string, as map[string]*analysis.Analyzer) (fields []string, err error) {
	for i, datum := range data {
		var doc *bluge.Document
		doc, fields, err = newDocument(datum, idField, storeFields, as)
		if err != nil {
			// todo warn or quit?
			return nil, errors.WithMessagef(err, "failed for item at index %d", i)
		}
		if err := iw.Insert(doc); err != nil {
			return nil, err
		}
	}
	return lo.Uniq(fields), nil
}

func newDocument(datum map[string]any, idField string, storeFields []string, as map[string]*analysis.Analyzer) (doc *bluge.Document, fields []string, err error) {
	id, ok := datum[idField]
	if !ok {
		return nil, nil, fmt.Errorf("id field %q not found in data item", idField)
	}
	doc = bluge.NewDocument(fmt.Sprint(id))
	for key, value := range datum {
		a, ok := as[key]
		if !ok {
			a = as["*"]
		}
		added := addField(doc, key, value, a, storeFields)
		if added == nil {
			// todo handle field errors
			continue
		}
		fields = append(fields, added...)
	}
	// add a composite field in order to search all fields if needed
	field := bluge.NewCompositeFieldExcluding("_all", []string{"_id", idField})
	a, ok := as["_all"]
	if !ok {
		a = as["*"]
	}
	field.WithAnalyzer(a)
	doc.AddField(field)
	return doc, fields, nil
}

func addField(doc *bluge.Document, key string, value interface{}, a bluge.Analyzer, storeFields []string) (fields []string) {
	if value == nil {
		return nil
	}
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.String:
		if fmt.Sprint(value) == "" {
			return nil
		}
		field := bluge.NewTextField(key, fmt.Sprint(value))
		addTermField(doc, field, a, storeFields)
		fields = append(fields, key)
	case reflect.Int:
		field := bluge.NewNumericField(key, value.(float64))
		addTermField(doc, field, a, storeFields)
		fields = append(fields, key)
	case reflect.Float64:
		field := bluge.NewNumericField(key, value.(float64))
		addTermField(doc, field, a, storeFields)
		fields = append(fields, key)
	case reflect.Bool:
		field := bluge.NewKeywordField(key, strconv.FormatBool(value.(bool)))
		addTermField(doc, field, a, storeFields)
		fields = append(fields, key)
	case reflect.Map:
		vm, ok := value.(map[string]any)
		if !ok {
			// todo handle other than map[string]any
			return nil
		}
		for k, v := range vm {
			fields = append(fields, addField(doc, fmt.Sprintf("%v.%v", key, k), v, a, storeFields)...)
		}
	case reflect.Slice:
		vs, ok := value.([]any)
		if !ok {
			return nil
		}
		for i, v := range vs {
			fields = append(fields, addField(doc, fmt.Sprintf("%v.[%v]", key, i), v, a, storeFields)...)
		}
	}
	return fields
}

func addTermField(d *bluge.Document, f *bluge.TermField, a bluge.Analyzer, storeFields []string) {
	if slices.Contains(storeFields, f.Name()) || slices.Contains(storeFields, "*") {
		f.StoreValue()
	}
	if a != nil {
		f.WithAnalyzer(a)
	}
	d.AddField(f)
}

func loadData(path string) ([]map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var data []map[string]interface{}
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
