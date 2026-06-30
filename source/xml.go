package source

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// XMLSource — стрим-парсер XML. ItemElement — имя элемента-«записи».
// Дочерние элементы становятся ключами RawRecord (атрибуты — с префиксом "@").
// Source не владеет переданным io.Reader и не закрывает его — это остаётся
// ответственностью вызывающего кода.
type XMLSource struct {
	ItemElement string

	decoder *xml.Decoder
}

func NewXML(item string) *XMLSource { return &XMLSource{ItemElement: item} }

func (s *XMLSource) Open(_ context.Context, r io.Reader) error {
	if s.ItemElement == "" {
		return fmt.Errorf("xml: ItemElement required")
	}
	s.decoder = xml.NewDecoder(r)
	return nil
}

func (s *XMLSource) Next(ctx context.Context) (RawRecord, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tok, err := s.decoder.Token()
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != s.ItemElement {
			continue
		}
		return s.decodeItem(se)
	}
}

func (s *XMLSource) decodeItem(start xml.StartElement) (RawRecord, error) {
	out := RawRecord{}
	for _, a := range start.Attr {
		out["@"+a.Name.Local] = a.Value
	}
	depth := 1
	var path []string
	var sb strings.Builder
	for depth > 0 {
		tok, err := s.decoder.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			path = append(path, t.Name.Local)
			sb.Reset()
			for _, a := range t.Attr {
				key := strings.Join(path, ".") + "@" + a.Name.Local
				out[key] = a.Value
			}
			depth++
		case xml.EndElement:
			if len(path) > 0 {
				key := strings.Join(path, ".")
				val := strings.TrimSpace(sb.String())
				if val != "" {
					if existing, ok := out[key]; ok {
						out[key] = existing + "|" + val
					} else {
						out[key] = val
					}
				}
				path = path[:len(path)-1]
				sb.Reset()
			}
			depth--
		case xml.CharData:
			sb.Write(t)
		}
	}
	return out, nil
}

func (s *XMLSource) Close() error {
	return nil
}
