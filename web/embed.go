package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed templates static
var content embed.FS

func ParseTemplates() (*template.Template, error) {
	tmpl, err := template.ParseFS(content, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded templates: %w", err)
	}

	return tmpl, nil
}

func StaticFiles() (http.FileSystem, error) {
	sub, err := fs.Sub(content, "static")
	if err != nil {
		return nil, fmt.Errorf("failed to create sub FileSystem for static: %w", err)
	}
	return http.FS(sub), nil
}
