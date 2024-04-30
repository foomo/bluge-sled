package sled

import (
	"encoding/json"
	"os"

	"github.com/brianvoe/gofakeit/v7"
)

func GenerateFakeData(num int, savePath string) ([]map[string]any, error) {
	var data []map[string]any
	for range num {
		data = append(data, map[string]any{
			"id":    gofakeit.UUID(),
			"image": gofakeit.URL(),
			"title": gofakeit.ProductName(),
			"color": gofakeit.Color(),
			"brand": gofakeit.Company(),
			"description": []string{
				gofakeit.ProductDescription(),
				gofakeit.ProductDescription(),
				gofakeit.ProductDescription(),
			},
			// "match": fmt.Sprintf("all %v", gofakeit.Name()), // in order to be able to test perf
		})
	}
	if savePath != "" {
		f, err := os.Create(savePath)
		if err != nil {
			return data, err
		}
		if err := json.NewEncoder(f).Encode(data); err != nil {
			return data, err
		}
	}
	return data, nil
}
