package server

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

var ErrNoPages = errors.New("No page has been found")

func parseTemplates(fsys fs.FS) (map[string]*template.Template, error) {
	subFS, err := fs.Sub(fsys, "templates")
	if err != nil {
		return nil, fmt.Errorf("read templates filesystem: %w", err)
	}

	pages, err := fs.Glob(subFS, "pages/*.html")
	if err != nil {
		return nil, fmt.Errorf("read templates filesystem: %w", err)
	}

	if len(pages) == 0 {
		return nil, ErrNoPages
	}

	pagesMap := make(map[string]*template.Template, len(pages))

	for _, page := range pages {
		fileName := path.Base(page)
		name := strings.TrimSuffix(fileName, ".html")

		base := template.New(name).Funcs(template.FuncMap{})

		pagesMap[name], err = base.ParseFS(subFS, "layouts/base.html", page)
		if err != nil {
			return nil, fmt.Errorf("Error while reading the layouts: %w", err)
		}
	}

	return pagesMap, nil
}

func (s *Server) render(w http.ResponseWriter, tmpName string, data any) {
	var buffer bytes.Buffer

	tmpl, ok := s.templates[tmpName]
	if !ok {
		s.logger.Error("template not found", "template", tmpName)
		http.Error(w, ErrNotFound.Error(), http.StatusInternalServerError)
		return
	}

	err := tmpl.ExecuteTemplate(&buffer, "base.html", data)
	if err != nil {
		s.logger.Error("failed to render template", "template", tmpName, "error", err)
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := buffer.WriteTo(w); err != nil {
		s.logger.Error("failed to write template response", "template", tmpName, "error", err)
	}
}
