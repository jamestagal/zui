package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

type Component struct {
	Name           string
	Data           map[string]interface{}
	OnMount        func() (func(), error)
	BeforeUpdate   func()
	AfterUpdate    func()
	Cleanup        func()
	mutex          sync.Mutex
	isMounted      bool
	updateCounter  int
	updateChan     chan struct{}
	pendingUpdates int
}

func NewComponent(name string) *Component {
	return &Component{
		Name:       name,
		Data:       make(map[string]interface{}),
		updateChan: make(chan struct{}, 100), // Buffer for pending updates
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
	go c.updateLoop() // Start the update loop
	return nil
}

func (c *Component) updateLoop() {
	for range c.updateChan {
		c.mutex.Lock()
		if c.BeforeUpdate != nil {
			c.BeforeUpdate()
		}
		c.updateCounter++
		if c.AfterUpdate != nil {
			c.AfterUpdate()
		}
		c.pendingUpdates--
		c.mutex.Unlock()
	}
}

func (c *Component) Update() {
	c.mutex.Lock()
	c.pendingUpdates++
	c.mutex.Unlock()
	c.updateChan <- struct{}{}
}

func (c *Component) Tick() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			c.mutex.Lock()
			if c.pendingUpdates == 0 {
				c.mutex.Unlock()
				close(done)
				return
			}
			c.mutex.Unlock()
			time.Sleep(time.Millisecond) // Small delay to prevent tight loop
		}
	}()
	return done
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

	close(c.updateChan)
	c.isMounted = false
}

func (c *Component) Render() template.HTML {
	return template.HTML(fmt.Sprintf("<div>Component: %s (Updates: %d)</div>", c.Name, c.updateCounter))
}

func handler(w http.ResponseWriter, r *http.Request) {
	comp := NewComponent("MyComponent")

	comp.OnMount = func() (func(), error) {
		log.Println("Component mounted!")
		return func() {
			log.Println("Component unmounted!")
		}, nil
	}

	comp.BeforeUpdate = func() {
		log.Println("Before update")
	}

	comp.AfterUpdate = func() {
		log.Println("After update")
	}

	err := comp.Mount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Simulate a few updates
	for i := 0; i < 3; i++ {
		comp.Update()
	}

	// Demonstrate tick functionality
	log.Println("Waiting for updates to complete...")
	<-comp.Tick()
	log.Println("All updates completed")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := `
    <html>
    <body>
        <h1>Svelte Lifecycle Demo (with Tick)</h1>
        {{.Component}}
        <p>Check the server console for lifecycle and tick messages.</p>
    </body>
    </html>
    `
	t := template.Must(template.New("page").Parse(tmpl))
	t.Execute(w, struct {
		Component template.HTML
	}{
		Component: comp.Render(),
	})

	defer comp.Unmount()
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Server is running on http://localhost:8084")
	http.ListenAndServe(":8084", nil)
}
