package importkit_test

import (
	"context"
	"strings"
	"testing"

	"github.com/laschenkov67/importkit"
)

func TestCSVImport_HappyPath(t *testing.T) {
	cfg := &importkit.Config{
		Format: importkit.FormatCSV,
		Source: importkit.SourceOptions{Delimiter: ";", HasHeader: true},
		Mappings: []importkit.FieldMapping{
			{Source: "sku", Target: "sku", Type: importkit.TypeString,
				Required: true, Transform: []string{"trim", "upper"}},
			{Source: "price", Target: "price", Type: importkit.TypeFloat,
				Validate: []string{"min=0"}},
			{Source: "qty", Target: "qty", Type: importkit.TypeInt, Default: 0},
		},
	}
	imp, err := importkit.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	in := strings.NewReader("sku;price;qty\n abc-1 ;199,90;5\nxyz;0;\n")
	recs, errs := imp.ImportAll(context.Background(), in)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(recs) != 2 {
		t.Fatalf("want 2 records, got %d", len(recs))
	}
	if recs[0]["sku"] != "ABC-1" {
		t.Errorf("transforms not applied: %v", recs[0]["sku"])
	}
	if recs[0]["price"].(float64) != 199.90 {
		t.Errorf("price coerce failed: %v", recs[0]["price"])
	}
	if recs[1]["qty"].(int64) != 0 {
		t.Errorf("default not applied: %v", recs[1]["qty"])
	}
}

func TestCSV_RequiredFieldMissing(t *testing.T) {
	cfg := &importkit.Config{
		Format: importkit.FormatCSV,
		Source: importkit.SourceOptions{Delimiter: ",", HasHeader: true},
		Mappings: []importkit.FieldMapping{
			{Source: "id", Target: "id", Type: importkit.TypeString, Required: true},
		},
	}
	imp, _ := importkit.New(cfg)
	// encoding/csv пропускает полностью пустые строки, поэтому используем
	// дополнительную колонку, чтобы строка не была пустой, а id — пустым.
	_, errs := imp.ImportAll(context.Background(), strings.NewReader("id,extra\n,present\n"))
	if len(errs) == 0 {
		t.Fatal("expected required error")
	}
}

func TestValidate_RegexFail(t *testing.T) {
	cfg := &importkit.Config{
		Format: importkit.FormatCSV,
		Source: importkit.SourceOptions{Delimiter: ",", HasHeader: true},
		Mappings: []importkit.FieldMapping{
			{Source: "email", Target: "email", Type: importkit.TypeString,
				Validate: []string{`regex=^[^@]+@[^@]+$`}},
		},
	}
	imp, _ := importkit.New(cfg)
	_, errs := imp.ImportAll(context.Background(),
		strings.NewReader("email\nbad-email\nok@example.com\n"))
	if len(errs) != 1 {
		t.Fatalf("want 1 err, got %d", len(errs))
	}
}
