package slack

import (
	"encoding/json"
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
			}).
			ParseFS(Templates, "templates/access-request.json.tmpl"),
	)
)
