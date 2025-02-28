package readme

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/yuin/goldmark"
)

// htmlTemplate is a simple HTML template that wraps the converted Markdown with dark styling.
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>README</title>
	<style>
		body {
			background-color: #121212;
			color: #e0e0e0;
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			margin: 2rem;
			line-height: 1.6;
		}
		pre {
			background-color: #1e1e1e;
			padding: 1rem;
			border-radius: 5px;
			overflow-x: auto;
		}
		code {
			font-family: "Courier New", Courier, monospace;
		}
		a {
			color: #82aaff;
		}
		h1, h2, h3, h4, h5, h6 {
			color: #ffffff;
		}
	</style>
</head>
<body>
	<div id="content">
		{{.Content}}
	</div>
</body>
</html>`

// ServeReadmeAsHTML reads README.md from the project root, converts it to HTML, and serves it with dark styling.
func ServeReadmeAsHTML(readmePath, port string) {
	content, err := ioutil.ReadFile(readmePath)
	if err != nil {
		log.Fatalf("Error reading %s: %v", readmePath, err)
	}

	// Convert Markdown to HTML using goldmark.
	var buf bytes.Buffer
	md := goldmark.New()
	if err := md.Convert(content, &buf); err != nil {
		log.Fatalf("Error converting Markdown to HTML: %v", err)
	}

	// Prepare the full HTML by wrapping the converted content with our template.
	tmpl, err := template.New("readme").Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("Error parsing HTML template: %v", err)
	}

	var fullHTML bytes.Buffer
	err = tmpl.Execute(&fullHTML, map[string]interface{}{
		"Content": buf.String(),
	})
	if err != nil {
		log.Fatalf("Error executing HTML template: %v", err)
	}

	// Set up the HTTP server.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(fullHTML.Bytes())
	})

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Serving %s as HTML on http://localhost:%s", readmePath, port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
