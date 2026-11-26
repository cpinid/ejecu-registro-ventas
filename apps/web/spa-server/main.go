package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	u, e := url.Parse(os.Getenv("BACKEND_URL"))
	if e != nil || u.Host == "" {
		log.Fatal("BACKEND_URL invalida")
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	mux := http.NewServeMux()
	mux.Handle("/api/", proxy)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join("/srv", filepath.Clean("/"+r.URL.Path))
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		if _, e := os.Stat(p); e != nil {
			p = "/srv/index.html"
		}
		http.ServeFile(w, r, p)
	})
	s := http.Server{Addr: ":8080", ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		mux.ServeHTTP(w, r)
	})}
	log.Fatal(s.ListenAndServe())
}
