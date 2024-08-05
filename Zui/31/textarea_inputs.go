package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type TextArea struct {
	Name  string
	Value string
	Rows  int
	Cols  int
}

func (ta *TextArea) Render() template.HTML {
	return template.HTML(fmt.Sprintf(`<textarea name="%s" rows="%d" cols="%d" id="biography">%s</textarea>`, ta.Name, ta.Rows, ta.Cols, ta.Value))
}

func handler(w http.ResponseWriter, r *http.Request) {
	textArea := &TextArea{
		Name:  "biography",
		Value: "",
		Rows:  4,
		Cols:  50,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := `
    <html>
    <head>
        <script>
            document.addEventListener('DOMContentLoaded', function() {
                const textarea = document.getElementById('biography');
                const display = document.getElementById('biographyDisplay');
                
                textarea.addEventListener('input', function() {
                    display.textContent = this.value;
                });
            });
        </script>
    </head>
    <body>
        <h1>Enter your biography:</h1>
        {{.TextArea}}
        <h2>Your biography:</h2>
        <p id="biographyDisplay"></p>
    </body>
    </html>
    `
	t := template.Must(template.New("page").Parse(tmpl))
	t.Execute(w, struct {
		TextArea template.HTML
	}{
		TextArea: textArea.Render(),
	})
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server is running on http://localhost:8082")
	http.ListenAndServe(":8082", nil)
}
