package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed templates
var content embed.FS

func TemplatesFS() http.FileSystem {
	subFS, err := fs.Sub(content, "templates")
	if err != nil {
		log.Fatal("web: couldn't find 'templates' directory in embedded files:", err)
	}
	return http.FS(subFS)
}