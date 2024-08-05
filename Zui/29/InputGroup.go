package main

import (
	"fmt"
	"net/http"
	"strings"
)

type InputGroup struct {
	Type    string
	Name    string
	Options []string
	Value   interface{}
}

func (ig *InputGroup) Bind(value string) {
	switch ig.Type {
	case "radio":
		ig.Value = value
	case "checkbox":
		values, ok := ig.Value.([]string)
		if !ok {
			values = []string{}
		}
		index := indexOf(values, value)
		if index == -1 {
			values = append(values, value)
		} else {
			values = append(values[:index], values[index+1:]...)
		}
		ig.Value = values
	}
}

func (ig *InputGroup) IsChecked(value string) bool {
	switch ig.Type {
	case "radio":
		return ig.Value == value
	case "checkbox":
		values, ok := ig.Value.([]string)
		if !ok {
			return false
		}
		return indexOf(values, value) != -1
	}
	return false
}

func (ig *InputGroup) Render() string {
	var sb strings.Builder
	for _, option := range ig.Options {
		checked := ""
		if ig.IsChecked(option) {
			checked = "checked"
		}
		sb.WriteString(fmt.Sprintf("<input type=\"%s\" name=\"%s\" value=\"%s\" %s />\n", ig.Type, ig.Name, option, checked))
	}
	return sb.String()
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

func handler(w http.ResponseWriter, r *http.Request) {
	tortilla := &InputGroup{
		Type:    "radio",
		Name:    "tortilla",
		Options: []string{"Plain", "Whole wheat", "Spinach"},
		Value:   "",
	}

	fillings := &InputGroup{
		Type:    "checkbox",
		Name:    "fillings",
		Options: []string{"Rice", "Beans", "Cheese", "Guac (extra)"},
		Value:   []string{},
	}

	// Simulate some bindings
	tortilla.Bind("Whole wheat")
	fillings.Bind("Rice")
	fillings.Bind("Cheese")

	fmt.Fprintf(w, "<h1>Tortilla:</h1>")
	fmt.Fprintf(w, tortilla.Render())
	fmt.Fprintf(w, "<h1>Fillings:</h1>")
	fmt.Fprintf(w, fillings.Render())
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
