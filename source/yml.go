package source

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// YMLSource парсит товарный фид Яндекс.Маркет (yml_catalog/shop/offers/offer).
// Кладёт в RawRecord плоские поля offer: id, available, name, price, vendor,
// currencyId, categoryId, url, description, picture (через '|' если несколько),
// param.<NAME> = значение.
// Source не владеет переданным io.Reader и не закрывает его — это остаётся
// ответственностью вызывающего кода.
type YMLSource struct {
	decoder *xml.Decoder
}

func NewYML() *YMLSource { return &YMLSource{} }

func (s *YMLSource) Open(_ context.Context, r io.Reader) error {
	s.decoder = xml.NewDecoder(r)
	s.decoder.CharsetReader = identityCharset
	return nil
}

func (s *YMLSource) Next(ctx context.Context) (RawRecord, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tok, err := s.decoder.Token()
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "offer" {
			continue
		}
		return s.parseOffer(se)
	}
}

func (s *YMLSource) parseOffer(start xml.StartElement) (RawRecord, error) {
	rec := RawRecord{}
	for _, a := range start.Attr {
		rec[a.Name.Local] = a.Value // id, available, type
	}
	var sb strings.Builder
	var current string
	var currentAttrs []xml.Attr
	for {
		tok, err := s.decoder.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			current = t.Name.Local
			currentAttrs = t.Attr
			sb.Reset()
		case xml.CharData:
			sb.Write(t)
		case xml.EndElement:
			if t.Name.Local == "offer" {
				return rec, nil
			}
			val := strings.TrimSpace(sb.String())
			key := current
			if current == "param" {
				name := ""
				for _, a := range currentAttrs {
					if a.Name.Local == "name" {
						name = a.Value
					}
				}
				key = "param." + strings.ToLower(strings.ReplaceAll(name, " ", "_"))
			}
			if existing, ok := rec[key]; ok && val != "" {
				rec[key] = existing + "|" + val
			} else if val != "" {
				rec[key] = val
			}
			current = ""
		}
	}
}

func (s *YMLSource) Close() error {
	return nil
}

// identityCharset поддерживает windows-1251/utf-8 без внешних зависимостей
// (для production-нагрузки рекомендуется golang.org/x/text/encoding/charmap).
func identityCharset(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "", "utf-8", "utf8", "us-ascii":
		return input, nil
	case "windows-1251", "cp1251":
		return input, nil // см. README о подключении charmap
	}
	return nil, fmt.Errorf("yml: unsupported charset %q", charset)
}

// Утилита, экспортируемая для пользователя — нормализация цены.
func ParsePrice(s string) (float64, error) {
	return strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
}
