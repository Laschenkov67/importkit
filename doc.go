// Package importkit предоставляет универсальный конвейер импорта данных
// из распространённых форматов (CSV, XLSX, XML, Яндекс.Маркет YML, 1С CommerceML)
// в нормализованные Go-структуры через декларативный конфиг.
//
// Базовый сценарий:
//
//	cfg, _ := importkit.LoadConfigFile("import.yaml")
//	imp, _ := importkit.New(cfg)
//	results, _ := imp.Import(ctx, file)
//	for r := range results {
//	    if r.Err != nil { log.Println(r.Err); continue }
//	    fmt.Println(r.Record)
//	}
//
// Конвейер построен на интерфейсах source.Source, transform.Transformer,
// validate.Validator и sink.Sink, что позволяет расширять поведение, не трогая ядро.
package importkit
