package api

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

//go:embed docs/*
var docsFS embed.FS

// TemplateData holds data for template rendering
type TemplateData struct {
	BaseURL string
}

// HandleOpenAPISpec serves the embedded OpenAPI specification
func (s *Server) HandleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("docs/openapi.yaml")
	if err != nil {
		http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Write(data)
}

// HandleAPIDocs serves the embedded API documentation HTML page
func (s *Server) HandleAPIDocs(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("docs/index.html")
	if err != nil {
		http.Error(w, "API documentation not found", http.StatusNotFound)
		return
	}

	// Parse the HTML template
	tmpl, err := template.New("docs").Parse(string(data))
	if err != nil {
		http.Error(w, "Failed to parse documentation template", http.StatusInternalServerError)
		return
	}

	// Determine base URL
	baseURL := getBaseURL(r)

	// Prepare template data
	templateData := TemplateData{
		BaseURL: baseURL,
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes

	if err := tmpl.Execute(w, templateData); err != nil {
		http.Error(w, "Failed to render documentation", http.StatusInternalServerError)
		return
	}
}

// HandleSwaggerUI serves the Swagger UI interface
func (s *Server) HandleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	data, err := docsFS.ReadFile("docs/swagger.html")
	if err != nil {
		http.Error(w, "Swagger UI not found", http.StatusNotFound)
		return
	}

	// Parse the HTML template
	tmpl, err := template.New("swagger").Parse(string(data))
	if err != nil {
		http.Error(w, "Failed to parse Swagger UI template", http.StatusInternalServerError)
		return
	}

	// Determine base URL
	baseURL := getBaseURL(r)

	// Prepare template data
	templateData := TemplateData{
		BaseURL: baseURL,
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes

	if err := tmpl.Execute(w, templateData); err != nil {
		http.Error(w, "Failed to render Swagger UI", http.StatusInternalServerError)
		return
	}
}

// getBaseURL constructs the base URL from the request
func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := r.Host
	if host == "" {
		host = r.Header.Get("X-Forwarded-Host")
	}
	if host == "" {
		host = "localhost"
	}

	// Remove any trailing slash
	baseURL := fmt.Sprintf("%s://%s", scheme, host)
	return strings.TrimSuffix(baseURL, "/")
}
