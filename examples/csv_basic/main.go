package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/laschenkov67/importkit"
)

func main() {
	cfg := &importkit.Config{
		Format: importkit.FormatCSV,
		Source: importkit.SourceOptions{Delimiter: ";", HasHeader: true},
		Mappings: []importkit.FieldMapping{
			{Source: "Артикул", Target: "sku", Type: importkit.TypeString,
				Required: true, Transform: []string{"trim", "upper"}},
			{Source: "Название", Target: "name", Type: importkit.TypeString,
				Required: true, Transform: []string{"trim"},
				Validate: []string{"len_max=255"}},
			{Source: "Цена", Target: "price", Type: importkit.TypeFloat,
				Validate: []string{"min=0"}},
			{Source: "Дата", Target: "imported_at", Type: importkit.TypeDate,
				Format: "02.01.2006"},
		},
	}
	imp, err := importkit.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("products.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	ch, err := imp.Import(context.Background(), f)
	if err != nil {
		log.Fatal(err)
	}
	for r := range ch {
		if r.Err != nil {
			log.Printf("row %d err: %v", r.Row, r.Err)
			continue
		}
		fmt.Printf("%d: %v\n", r.Row, r.Record)
	}
}
