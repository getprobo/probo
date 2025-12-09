package slack

import (
	"encoding/json"
	"strings"
	"text/template"
)

var (
	accessRequestTemplate = template.Must(
		template.New("access-request.json.tmpl").
			Funcs(template.FuncMap{
				"jsonEscape": func(s string) string {
					b, _ := json.Marshal(s)
					return string(b[1 : len(b)-1])
				},
				"buildAcceptAllValue": func(docIDs, repIDs []string) string {
					value := map[string][]string{
						"document_ids": docIDs,
						"report_ids":   repIDs,
					}
					b, _ := json.Marshal(value)
					s := string(b)
					s = strings.ReplaceAll(s, `\`, `\\`)
					s = strings.ReplaceAll(s, `"`, `\"`)
					return s
				},
			}).
			ParseFS(Templates, "templates/access-request.json.tmpl"),
	)
)
