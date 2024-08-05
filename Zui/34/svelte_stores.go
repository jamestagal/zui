package main

import (
	"html/template"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

type Subscriber struct {
	id  int64
	run func(interface{})
}

type Store struct {
	value       interface{}
	subscribers map[int64]Subscriber
	mutex       sync.RWMutex
	nextID      int64
}

func NewStore(initialValue interface{}) *Store {
	return &Store{
		value:       initialValue,
		subscribers: make(map[int64]Subscriber),
	}
}

func (s *Store) Subscribe(run func(interface{})) func() {
	s.mutex.Lock()
	id := atomic.AddInt64(&s.nextID, 1)
	s.subscribers[id] = Subscriber{id, run}
	s.mutex.Unlock()

	run(s.value) // Immediately call the subscriber with current value

	return func() {
		s.mutex.Lock()
		delete(s.subscribers, id)
		s.mutex.Unlock()
	}
}

func (s *Store) Set(newValue interface{}) {
	s.mutex.Lock()
	s.value = newValue
	subscribers := make([]Subscriber, 0, len(s.subscribers))
	for _, sub := range s.subscribers {
		subscribers = append(subscribers, sub)
	}
	s.mutex.Unlock()

	for _, sub := range subscribers {
		sub.run(newValue)
	}
}

func (s *Store) Update(updater func(interface{}) interface{}) {
	s.mutex.Lock()
	newValue := updater(s.value)
	s.value = newValue
	subscribers := make([]Subscriber, 0, len(s.subscribers))
	for _, sub := range s.subscribers {
		subscribers = append(subscribers, sub)
	}
	s.mutex.Unlock()

	for _, sub := range subscribers {
		sub.run(newValue)
	}
}

type AppState struct {
	Count *Store
}

func NewAppState() *AppState {
	return &AppState{
		Count: NewStore(0),
	}
}

func handler(w http.ResponseWriter, r *http.Request, state *AppState) {
	var countValue interface{}
	unsubscribe := state.Count.Subscribe(func(value interface{}) {
		countValue = value
		log.Printf("Count updated: %v\n", value)
	})
	defer unsubscribe()

	if r.Method == "POST" {
		state.Count.Update(func(value interface{}) interface{} {
			return value.(int) + 1
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := `
    <html>
    <body>
        <h1>Svelte Store Demo</h1>
        <p>Count: {{.Count}}</p>
        <form method="post">
            <button type="submit">Increment</button>
        </form>
    </body>
    </html>
    `
	t := template.Must(template.New("page").Parse(tmpl))
	t.Execute(w, struct {
		Count interface{}
	}{
		Count: countValue,
	})
}

func main() {
	state := NewAppState()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, state)
	})

	log.Println("Server is running on http://localhost:8085")
	http.ListenAndServe(":8085", nil)
}
