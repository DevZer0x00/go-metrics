package assets

import (
	"embed"
	"html/template"
	"io"
)

//go:embed html/**/*.tmpl
var templatesFS embed.FS
var templates *template.Template

type GetAlMetricsTemplateData struct {
	Name  string
	Value string
}

func getTemplates() (*template.Template, error) {
	if templates == nil {
		tmpls, err := template.ParseFS(templatesFS, "html/**/*.tmpl")
		if err != nil {
			return nil, err
		}

		templates = tmpls
	}

	return templates, nil
}

func ExecuteGetAllMetricsTemplate(w io.Writer, data []*GetAlMetricsTemplateData) error {
	tmpls, err := getTemplates()
	if err != nil {
		return err
	}

	return tmpls.ExecuteTemplate(w, "metrics_list.tmpl", data)
}
