package readme

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// htmlTemplate is an enhanced HTML template with modern dark mode styling
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>README</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/styles/atom-one-dark.min.css">
    <style>
        :root {
            --bg-primary: #0d1117;
            --bg-secondary: #161b22;
            --bg-tertiary: #21262d;
            --text-primary: #e6edf3;
            --text-secondary: #c9d1d9;
            --text-muted: #8b949e;
            --border-color: #30363d;
            --accent-color: #58a6ff;
            --accent-hover: #79c0ff;
            --success-color: #3fb950;
            --warning-color: #d29922;
            --error-color: #f85149;
            --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji';
            --font-mono: SFMono-Regular, Consolas, 'Liberation Mono', Menlo, monospace;
            --max-width: 960px;
            --radius-sm: 4px;
            --radius-md: 6px;
            --radius-lg: 8px;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        html, body {
            height: 100%;
            width: 100%;
        }

        body {
            background-color: var(--bg-primary);
            color: var(--text-primary);
            font-family: var(--font-sans);
            font-size: 16px;
            line-height: 1.6;
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
            text-rendering: optimizeLegibility;
        }

        #container {
            max-width: var(--max-width);
            margin: 0 auto;
            padding: 2rem;
        }

        #content {
            background-color: var(--bg-secondary);
            border-radius: var(--radius-lg);
            padding: 2rem;
            border: 1px solid var(--border-color);
            box-shadow: 0 4px 24px rgba(0, 0, 0, 0.25);
        }

        /* Header styles */
        h1, h2, h3, h4, h5, h6 {
            color: var(--text-primary);
            font-weight: 600;
            margin: 1.5em 0 0.75em 0;
            position: relative;
        }

        h1 {
            font-size: 2rem;
            padding-bottom: 0.5rem;
            border-bottom: 1px solid var(--border-color);
            margin-top: 0;
        }

        h2 {
            font-size: 1.5rem;
            padding-bottom: 0.3rem;
            border-bottom: 1px solid var(--border-color);
        }

        h3 { font-size: 1.25rem; }
        h4 { font-size: 1rem; }
        h5 { font-size: 0.875rem; }
        h6 { font-size: 0.85rem; }

        /* Link anchor tags for headings */
        h1:hover::before, h2:hover::before, h3:hover::before, 
        h4:hover::before, h5:hover::before, h6:hover::before {
            content: "#";
            position: absolute;
            left: -1.25rem;
            color: var(--accent-color);
            font-weight: normal;
            opacity: 0.6;
        }

        /* Text elements */
        p {
            margin: 0 0 1rem 0;
            color: var(--text-secondary);
        }

        a {
            color: var(--accent-color);
            text-decoration: none;
            transition: color 0.2s ease;
        }

        a:hover {
            color: var(--accent-hover);
            text-decoration: underline;
        }

        strong {
            font-weight: 600;
        }

        em {
            font-style: italic;
        }

        blockquote {
            margin: 1rem 0;
            padding: 0.5rem 1rem;
            border-left: 4px solid var(--accent-color);
            background-color: rgba(88, 166, 255, 0.1);
            border-radius: var(--radius-sm);
        }

        blockquote > p {
            margin-bottom: 0;
        }

        /* Lists */
        ul, ol {
            margin: 1rem 0;
            padding-left: 2rem;
            color: var(--text-secondary);
        }

        li {
            margin: 0.25rem 0;
        }

        li > ul, li > ol {
            margin: 0.25rem 0;
        }

        /* Code blocks and inline code */
        pre {
            background-color: var(--bg-tertiary);
            border-radius: var(--radius-md);
            padding: 1rem;
            overflow-x: auto;
            margin: 1rem 0;
            border: 1px solid var(--border-color);
            position: relative;
            font-family: var(--font-mono);
            font-size: 0.875rem;
        }

        pre::before {
            content: "";
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 28px;
            background: rgba(255, 255, 255, 0.05);
            border-bottom: 1px solid var(--border-color);
            border-top-left-radius: var(--radius-md);
            border-top-right-radius: var(--radius-md);
            z-index: 0;
        }

        pre::after {
            content: "";
            position: absolute;
            top: 8px;
            left: 10px;
            height: 12px;
            width: 12px;
            border-radius: 50%;
            background-color: var(--error-color);
            box-shadow: 25px 0 0 var(--warning-color), 50px 0 0 var(--success-color);
            z-index: 1;
        }

        pre code {
            padding: 0;
            background-color: transparent;
            border-radius: 0;
            font-family: inherit;
            position: relative;
            top: 14px;
            display: block;
        }

        code {
            font-family: var(--font-mono);
            font-size: 0.875em;
            background-color: rgba(110, 118, 129, 0.1);
            padding: 0.2em 0.4em;
            border-radius: var(--radius-sm);
            color: var(--text-primary);
        }

        /* Tables */
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 1rem 0;
            border-radius: var(--radius-md);
            overflow: hidden;
            border: 1px solid var(--border-color);
        }

        thead {
            background-color: var(--bg-tertiary);
        }

        th, td {
            padding: 0.75rem;
            border: 1px solid var(--border-color);
            text-align: left;
        }

        th {
            font-weight: 600;
            color: var(--text-primary);
        }

        tr:nth-child(even) {
            background-color: rgba(255, 255, 255, 0.02);
        }

        /* Images */
        img {
            max-width: 100%;
            height: auto;
            border-radius: var(--radius-md);
            border: 1px solid var(--border-color);
            display: block;
            margin: 1.5rem auto;
        }

        /* Horizontal rule */
        hr {
            height: 1px;
            background-color: var(--border-color);
            border: none;
            margin: 2rem 0;
        }

        /* Custom container classes */
        .note, .tip, .warning, .danger {
            margin: 1rem 0;
            padding: 1rem;
            border-radius: var(--radius-md);
            border-left: 4px solid;
        }

        .note {
            background-color: rgba(88, 166, 255, 0.1);
            border-left-color: var(--accent-color);
        }

        .tip {
            background-color: rgba(63, 185, 80, 0.1);
            border-left-color: var(--success-color);
        }

        .warning {
            background-color: rgba(210, 153, 34, 0.1);
            border-left-color: var(--warning-color);
        }

        .danger {
            background-color: rgba(248, 81, 73, 0.1);
            border-left-color: var(--error-color);
        }

        /* Footer */
        .footer {
            margin-top: 2rem;
            padding-top: 1rem;
            border-top: 1px solid var(--border-color);
            color: var(--text-muted);
            font-size: 0.875rem;
            text-align: center;
        }

        /* Responsive adjustments */
        @media (max-width: 768px) {
            #container {
                padding: 1rem;
            }

            #content {
                padding: 1.5rem;
            }

            h1 {
                font-size: 1.75rem;
            }

            h2 {
                font-size: 1.25rem;
            }
        }
    </style>
</head>
<body>
    <div id="container">
        <div id="content">
            {{.Content}}
        </div>
        <div class="footer">
            <p>Generated with <i class="fas fa-heart"></i> using Go README Renderer</p>
        </div>
    </div>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/highlight.min.js"></script>
    <script>
        // Apply code highlighting
        document.addEventListener('DOMContentLoaded', (event) => {
            document.querySelectorAll('pre code').forEach((block) => {
                hljs.highlightElement(block);
            });

            // Convert h1-h6 to have anchor links
            document.querySelectorAll('h1, h2, h3, h4, h5, h6').forEach((heading) => {
                // Create anchor id from heading text
                const id = heading.textContent.toLowerCase().replace(/[^\w]+/g, '-');
                heading.setAttribute('id', id);
                
                // Make headings clickable to copy URL
                heading.addEventListener('click', () => {
                    const url = window.location.href.split('#')[0] + '#' + id;
                    navigator.clipboard.writeText(url);
                    
                    // Visual feedback
                    const originalColor = heading.style.color;
                    heading.style.color = 'var(--accent-color)';
                    setTimeout(() => {
                        heading.style.color = originalColor;
                    }, 300);
                });
            });
        });
    </script>
</body>
</html>`

// ServeReadmeAsHTML reads README.md from the project root, converts it to HTML, and serves it with modern dark styling.
func ServeReadmeAsHTML(readmePath, port string) {
	content, err := ioutil.ReadFile(readmePath)
	if err != nil {
		log.Fatalf("Error reading %s: %v", readmePath, err)
	}

	// Configure goldmark with GitHub Flavored Markdown extensions
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // Allows raw HTML in the markdown
		),
	)

	// Convert Markdown to HTML
	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		log.Fatalf("Error converting Markdown to HTML: %v", err)
	}

	// Prepare the full HTML by wrapping the converted content with our template
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

	// Set up the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(fullHTML.Bytes())
	})

	// Set up static file serving for potential assets
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("✨ Serving %s as HTML on http://localhost:%s", readmePath, port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
