package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type Component struct {
	Name      string
	Data      map[string]interface{}
	OnMount   func() (func(), error)
	Cleanup   func()
	mutex     sync.Mutex
	isMounted bool
}

func NewComponent(name string) *Component {
	return &Component{
		Name: name,
		Data: make(map[string]interface{}),
	}
}

func (c *Component) Mount() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isMounted {
		return nil
	}

	if c.OnMount != nil {
		cleanup, err := c.OnMount()
		if err != nil {
			return err
		}
		c.Cleanup = cleanup
	}

	c.isMounted = true
	return nil
}

func (c *Component) Unmount() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isMounted {
		return
	}

	if c.Cleanup != nil {
		c.Cleanup()
	}

	c.isMounted = false
}

func (c *Component) Render() template.HTML {
	return template.HTML(fmt.Sprintf("<div>Component: %s</div>", c.Name))
}

func handler(w http.ResponseWriter, r *http.Request) {
	comp := NewComponent("MyComponent")

	comp.OnMount = func() (func(), error) {
		log.Println("Component mounted!")
		return func() {
			log.Println("Component unmounted!")
		}, nil
	}

	err := comp.Mount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := `
    <html>
    <body>
        <h1>Svelte Lifecycle Demo</h1>
        {{.Component}}
        <p>Check the server console for mount/unmount messages.</p>
    </body>
    </html>
    `
	t := template.Must(template.New("page").Parse(tmpl))
	t.Execute(w, struct {
		Component template.HTML
	}{
		Component: comp.Render(),
	})

	// Simulate unmounting after rendering
	// In a real application, this would happen when the component is destroyed
	defer comp.Unmount()
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Server is running on http://localhost:8083")
	http.ListenAndServe(":8083", nil)
}
