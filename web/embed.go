package web

import (
	"embed"
	"html/template"
	"log"
)

//go:embed templates
var content embed.FS

func ParseTemplates() *template.Template {
	tmpl, err := template.ParseFS(content, "templates/*.html")
	if err != nil {
		log.Fatalf("couldn't parse embedded templates: %v", err)
	}

	return tmpl
}