package web

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// Templates manages HTML templates for the explorer web interface
type Templates struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewTemplates creates a new templates manager
func NewTemplates() *Templates {
	t := &Templates{
		templates: make(map[string]*template.Template),
		funcMap: template.FuncMap{
			"formatHash":      formatHash,
			"formatAddress":   formatAddress,
			"formatAmount":    formatAmount,
			"formatTime":      formatTime,
			"formatDifficulty": formatDifficulty,
			"add":             add,
			"sub":             sub,
			"mul":             mul,
			"div":             div,
			"mod":             mod,
		},
	}

	// Load all templates
	t.loadTemplates()
	return t
}

// Render renders a template with the given data
func (t *Templates) Render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, exists := t.templates[name]
	if !exists {
		http.Error(w, fmt.Sprintf("Template %s not found", name), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

// loadTemplates loads all HTML templates
func (t *Templates) loadTemplates() {
	templateFiles := []string{
		"base.html",
		"home.html",
		"blocks.html",
		"block_detail.html",
		"transactions.html",
		"transaction_detail.html",
		"address_detail.html",
		"search.html",
		"search_results.html",
		"error.html",
	}

	for _, filename := range templateFiles {
		t.loadTemplate(filename)
	}
}

// loadTemplate loads a single template
func (t *Templates) loadTemplate(filename string) {
	// For content templates, we need to parse both base and content together
	if filename != "base.html" {
		baseContent := t.getTemplateContent("base.html")
		contentContent := t.getTemplateContent(filename)
		
		// Parse base template first
		tmpl, err := template.New("base").Funcs(t.funcMap).Parse(baseContent)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse base template: %v", err))
		}
		
		// Parse content template into the same template
		_, err = tmpl.Parse(contentContent)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse content template %s: %v", filename, err))
		}
		
		t.templates[filename] = tmpl
	} else {
		// Base template is loaded separately for reference
		content := t.getTemplateContent(filename)
		tmpl, err := template.New(filename).Funcs(t.funcMap).Parse(content)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse template %s: %v", filename, err))
		}
		t.templates[filename] = tmpl
	}
}

// getTemplateContent returns the content for a template
func (t *Templates) getTemplateContent(filename string) string {
	switch filename {
	case "base.html":
		return baseTemplate
	case "home.html":
		return homeTemplate
	case "blocks.html":
		return blocksTemplate
	case "block_detail.html":
		return blockDetailTemplate
	case "transactions.html":
		return transactionsTemplate
	case "transaction_detail.html":
		return transactionDetailTemplate
	case "address_detail.html":
		return addressDetailTemplate
	case "search.html":
		return searchTemplate
	case "search_results.html":
		return searchResultsTemplate
	case "error.html":
		return errorTemplate
	default:
		return fmt.Sprintf("<!-- Template %s not found -->", filename)
	}
}

// Template helper functions

func formatHash(hash []byte) string {
	if len(hash) == 0 {
		return "N/A"
	}
	if len(hash) <= 8 {
		return fmt.Sprintf("%x", hash)
	}
	return fmt.Sprintf("%x...%x", hash[:4], hash[len(hash)-4:])
}

func formatAddress(address string) string {
	if len(address) <= 12 {
		return address
	}
	return fmt.Sprintf("%s...%s", address[:8], address[len(address)-4:])
}

func formatAmount(amount uint64) string {
	// Convert satoshis to a more readable format
	if amount == 0 {
		return "0"
	}
	
	// Assuming 8 decimal places (like Bitcoin)
	satoshis := float64(amount) / 100000000.0
	return fmt.Sprintf("%.8f", satoshis)
}

func formatTime(t interface{}) string {
	switch v := t.(type) {
	case int64:
		// Unix timestamp
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", t)
	}
}

func formatDifficulty(difficulty uint64) string {
	if difficulty == 0 {
		return "0"
	}
	
	// Format difficulty in a human-readable way
	if difficulty >= 1000000000 {
		return fmt.Sprintf("%.2f G", float64(difficulty)/1000000000.0)
	} else if difficulty >= 1000000 {
		return fmt.Sprintf("%.2f M", float64(difficulty)/1000000.0)
	} else if difficulty >= 1000 {
		return fmt.Sprintf("%.2f K", float64(difficulty)/1000.0)
	}
	return fmt.Sprintf("%d", difficulty)
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func mul(a, b int) int {
	return a * b
}

func div(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

func mod(a, b int) int {
	if b == 0 {
		return 0
	}
	return a % b
}
