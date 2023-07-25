package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"chatApp.azizrmadi.net/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("chat/templates", t.filename)))
	})

	data := make(map[string]interface{})
	data["Host"] = r.Host

	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	err := t.templ.Execute(w, data)
	if err != nil {
		log.Fatal("Parsing Error", err)
	}

}

func main() {
	addr := flag.String("addr", ":4000", "The addr of the application")
	traceOn := flag.Bool("trace", false, "Show trace logs in the app")
	flag.Parse()
	gomniauth.SetSecurityKey("Trying your best is a mindset not a choice")
	gomniauth.WithProviders(google.New("646781883218-bomsoqto4rn89pmrq6hqd5p7fghkfv9e.apps.googleusercontent.com", "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", "http://localhost:4000/auth/callback/google"))
	r := newRoom()
	if *traceOn {
		r.tracer = trace.New(os.Stdout)
	}
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)

	http.Handle("/room", r)
	go r.run()
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
