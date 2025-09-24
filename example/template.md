
{{range .Highlights}}
### {{.Timestamp}}

{{if .Text}}
> {{.Text}}
{{end}}

{{if .Note}}
**My Note:** {{.Note}}
{{end}}

{{end}}

