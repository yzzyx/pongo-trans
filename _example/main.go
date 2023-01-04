package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"

	"github.com/flosch/pongo2/v6"
	"github.com/yzzyx/pongo-trans"
)

// If the files should be loaded from an embedded FS, add this instead:
// //go:embed locales
// var localeFS embed.FS

func main() {
	// Read all translations from directory "locales"
	t, err := trans.NewTemplateTranslator(os.DirFS("locales"), ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create translator: %v\n", err)
		os.Exit(1)
	}

	// If the files should be loaded from an embedded FS, use this line instead
	// t, err := trans.NewTemplateTranslator(localeFS, "locales")

	pongo2.RegisterTag("trans", trans.NewTransTag(t))
	pongo2.RegisterTag("blocktrans", trans.NewBlockTransTag(t))

	tmpl, _ := pongo2.FromString(`{% trans "Please translate this!" %}`)
	result, _ := tmpl.Execute(pongo2.Context{
		"_language": "sv_SE",   // No default, uses string as-is if not specified
		"_domain":   "default", // Defaults to 'default', expecting a file called 'default.po' or 'default.mo'
	})

	fsLoader, err := pongo2.NewLocalFileSystemLoader("templates")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize template loader: %v\n", err)
		os.Exit(1)
	}

	ts := pongo2.NewSet("", fsLoader)

	var visitCount int32

	var srv http.Server

	srv.Addr = "127.0.0.1:9911"
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Normally this would probably be attached to a session or similar
		language := req.FormValue("lang")
		if language == "" {
			language = "en_GB"
		}
		vc := atomic.AddInt32(&visitCount, 1)

		tmpl, err := ts.FromFile("index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not load template: %v\n", err)
			return
		}

		err = tmpl.ExecuteWriter(pongo2.Context{
			"_language": language,
			"_domain":   "default",
			"p2":        "pongo2",
			"visited":   vc,
		}, w)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not render template: %v\n", err)
			return
		}
	})
	srv.Handler = mux

	// Cancel on ctrl+c
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	fmt.Printf("Starting server on http://%s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed

	fmt.Println("Translated result:", result)
}
