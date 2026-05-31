package source

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// OneCSource читает CommerceML 2.x: <КоммерческаяИнформация><Каталог><Товары><Товар>.
// Поля Товара превращаются в плоские ключи:
//
//	Ид, Артикул, Наименование, Описание, БазоваяЕдиница,
//	Группы.Ид (через '|'), ЗначенияСвойств.<Ид>, Картинка (через '|').
type OneCSource struct {
	decoder *xml.Decoder
	closer  io.Closer
}

func NewOneC() *OneCSource { return &OneCSource{} }

func (s *OneCSource) Open(_ context.Context, r io.Reader) error {
	s.decoder = xml.NewDecoder(r)
	s.decoder.CharsetReader = identityCharset
	if c, ok := r.(io.Closer); ok {
		s.closer = c
	}
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
			// Внутри <ЗначенияСвойства> структура: Ид + Значение.
			if name == "Ид" && len(path) >= 2 && path[len(path)-2] == "ЗначенияСвойства" {
				propID = val
			} else if name == "Значение" && len(path) >= 2 && path[len(path)-2] == "ЗначенияСвойства" {
				if propID != "" {
					appendVal("ЗначенияСвойств."+propID, val)
				}
			} else if name == "Ид" && len(path) >= 2 && path[len(path)-2] == "Группы" {
				appendVal("Группы.Ид", val)
			} else if name == "Картинка" {
				appendVal("Картинка", val)
			} else if val != "" {
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
	if s.closer != nil {
		return s.closer.Close()
	}
	return nil
}

var _ = fmt.Sprintf
