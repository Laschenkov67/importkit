package source

import (
	"context"
	"encoding/xml"
	"io"
	"strings"
)

// OneCSource читает CommerceML 2.x: <КоммерческаяИнформация><Каталог><Товары><Товар>.
// Поля Товара превращаются в плоские ключи:
//
//	Ид, Артикул, Наименование, Описание, БазоваяЕдиница,
//	Группы.Ид (через '|'), ЗначенияСвойств.<Ид>, Картинка (через '|').
//
// Source не владеет переданным io.Reader и не закрывает его — это остаётся
// ответственностью вызывающего кода.
type OneCSource struct {
	decoder *xml.Decoder
}

func NewOneC() *OneCSource { return &OneCSource{} }

func (s *OneCSource) Open(_ context.Context, r io.Reader) error {
	s.decoder = xml.NewDecoder(r)
	s.decoder.CharsetReader = identityCharset
	return nil
}

func (s *OneCSource) Next(ctx context.Context) (RawRecord, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tok, err := s.decoder.Token()
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "Товар" {
			continue
		}
		return s.parseProduct()
	}
}

func (s *OneCSource) parseProduct() (RawRecord, error) {
	rec := RawRecord{}
	depth := 1
	var path []string
	var sb strings.Builder
	var propID string

	appendVal := func(key, val string) {
		if val == "" {
			return
		}
		if existing, ok := rec[key]; ok {
			rec[key] = existing + "|" + val
		} else {
			rec[key] = val
		}
	}

	for depth > 0 {
		tok, err := s.decoder.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			path = append(path, t.Name.Local)
			sb.Reset()
			if t.Name.Local == "ЗначенияСвойства" {
				propID = ""
			}
			depth++
		case xml.CharData:
			sb.Write(t)
		case xml.EndElement:
			val := strings.TrimSpace(sb.String())
			name := t.Name.Local
			parent := ""
			if len(path) >= 2 {
				parent = path[len(path)-2]
			}
			// Внутри <ЗначенияСвойства> структура: Ид + Значение.
			switch {
			case name == "Ид" && parent == "ЗначенияСвойства":
				propID = val
			case name == "Значение" && parent == "ЗначенияСвойства":
				if propID != "" {
					appendVal("ЗначенияСвойств."+propID, val)
				}
			case name == "Ид" && parent == "Группы":
				appendVal("Группы.Ид", val)
			case name == "Картинка":
				appendVal("Картинка", val)
			case val != "":
				appendVal(name, val)
			}
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			sb.Reset()
			depth--
		}
	}
	return rec, nil
}

func (s *OneCSource) Close() error {
	return nil
}
