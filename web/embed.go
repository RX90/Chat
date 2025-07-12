package web

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

//go:embed templates static
var content embed.FS

func ParseTemplates() *template.Template {
	tmpl, err := template.ParseFS(content, "templates/*.html")
	if err != nil {
		log.Fatalf("couldn't parse embedded templates: %v", err)
	}

	return tmpl
}

func StaticFiles() http.FileSystem {
	sub, err := fs.Sub(content, "static")
	if err != nil {
		log.Fatalf("failed to create sub FS for static: %v", err)
	}
	return http.FS(sub)
}
