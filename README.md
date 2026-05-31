# importkit

Универсальный потоковый импортер данных для Go: CSV, XLSX, XML,
Яндекс.Маркет YML и 1С CommerceML — единый декларативный конфиг,
маппинг колонок, преобразования и валидация.

## Установка
```bash
go get github.com/laschenkov67/importkit
```

## Быстрый старт
```go
cfg := &amp;importkit.Config{
    Format: importkit.FormatCSV,
    Source: importkit.SourceOptions{Delimiter: ";", HasHeader: true},
    Mappings: []importkit.FieldMapping{
        {Source: "sku", Target: "sku", Type: importkit.TypeString,
            Required: true, Transform: []string{"trim", "upper"}},
        {Source: "price", Target: "price", Type: importkit.TypeFloat,
            Validate: []string{"min=0"}},
    },
}
imp, _ := importkit.New(cfg)
ch, _ := imp.Import(ctx, file)
for r := range ch {
    if r.Err != nil { log.Println(r.Err); continue }
    fmt.Println(r.Record)
}
```

## Расширение
Свои трансформеры:
```go
importkit.RegisterTransformer("slug", func(v any, _ string) (any, error) {
    return slug.Make(fmt.Sprint(v)), nil
})
```

## Форматы
| Формат | `Format`        | Особенности |
|--------|-----------------|-------------|
| CSV    | `csv`           | Любой разделитель, опциональный заголовок |
| XLSX   | `xlsx`          | Лист по имени, заголовок и DataStart |
| XML    | `xml`           | Стрим-парсер, `item_element` |
| YML    | `yml`           | Яндекс.Маркет offers, `param.<name>` |
| 1С     | `onec`          | CommerceML 2.x `<Товар>` |

## Лицензия
MIT