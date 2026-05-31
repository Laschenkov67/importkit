package source_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/laschenkov67/importkit/source"
)

const ymlSample = `<?xml version="1.0" encoding="UTF-8"?>
<yml_catalog date="2024-01-01"><shop><offers>
<offer id="100" available="true">
  <name>Книга про Go</name>
  <price>590.00</price>
  <currencyId>RUB</currencyId>
  <vendor>Acme</vendor>
  <param name="Автор">Иванов</param>
  <param name="Год">2024</param>
</offer>
<offer id="200" available="false">
  <name>Кепка</name><price>1200</price><currencyId>RUB</currencyId>
</offer>
</offers></shop></yml_catalog>`

func TestYML_Stream(t *testing.T) {
	s := source.NewYML()
	if err := s.Open(context.Background(), strings.NewReader(ymlSample)); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var cnt int
	for {
		rec, err := s.Next(context.Background())
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if rec["name"] == "" {
			t.Errorf("name missing: %v", rec)
		}
		if cnt == 0 && rec["param.автор"] != "Иванов" {
			t.Errorf("param parsed: %v", rec["param.автор"])
		}
		cnt++
	}
	if cnt != 2 {
		t.Errorf("want 2 offers, got %d", cnt)
	}
}
