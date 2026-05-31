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
		Format: importkit.FormatYML,
		Mappings: []importkit.FieldMapping{
			{Source: "id", Target: "external_id", Type: importkit.TypeString, Required: true},
			{Source: "available", Target: "in_stock", Type: importkit.TypeBool, Default: "false"},
			{Source: "name", Target: "title", Type: importkit.TypeString, Required: true,
				Transform: []string{"trim"}, Validate: []string{"len_min=2"}},
			{Source: "price", Target: "price", Type: importkit.TypeFloat,
				Validate: []string{"min=0"}},
			{Source: "currencyId", Target: "currency", Type: importkit.TypeString,
				Default: "RUB", Validate: []string{"in=RUB,USD,EUR"}},
			{Source: "vendor", Target: "brand", Type: importkit.TypeString,
				Transform: []string{"trim"}},
			{Source: "picture", Target: "images", Type: importkit.TypeString},
		},
	}
	imp, _ := importkit.New(cfg)
	f, err := os.Open("feed.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	ch, _ := imp.Import(context.Background(), f)
	for r := range ch {
		if r.Err != nil {
			log.Println(r.Err)
			continue
		}
		fmt.Println(r.Record["external_id"], r.Record["title"], r.Record["price"])
	}
}
