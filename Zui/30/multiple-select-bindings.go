package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type MultipleSelect struct {
	Name    string
	Options []string
	Value   []string
}

func (ms *MultipleSelect) Render() template.HTML {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<select multiple name=\"%s[]\" id=\"%s\">\n", ms.Name, ms.Name))
	for _, option := range ms.Options {
		sb.WriteString(fmt.Sprintf("  <option value=\"%s\">%s</option>\n", option, option))
	}
	sb.WriteString("</select>")
	return template.HTML(sb.String())
}

func handler(w http.ResponseWriter, r *http.Request) {
	fillings := &MultipleSelect{
		Name:    "fillings",
		Options: []string{"Rice", "Beans", "Cheese", "Guac (extra)"},
		Value:   []string{},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := `
    <html>
    <head>
        <script>
            document.addEventListener('DOMContentLoaded', function() {
                const select = document.getElementById('fillings');
                const display = document.getElementById('selectedFillings');
                
                select.addEventListener('change', function() {
                    const selectedOptions = Array.from(this.selectedOptions).map(option => option.value);
                    display.textContent = selectedOptions.join(', ');
                });
            });
        </script>
    </head>
    <body>
        <h1>Select your fillings:</h1>
        {{.Select}}
        <h2>Selected fillings:</h2>
        <p id="selectedFillings"></p>
    </body>
    </html>
    `
	t := template.Must(template.New("page").Parse(tmpl))
	t.Execute(w, struct {
		Select template.HTML
	}{
		Select: fillings.Render(),
	})
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server is running on http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}
