package middleware_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ar3s3ru/chi/middleware"
	"github.com/go-chi/chi"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func runServer(port int) {
	r := chi.NewRouter()
	r.Use(middleware.SelfDescribe())

	r.Get("/get", helloWorld)
	r.Post("/post", helloWorld)
	r.Put("/put", helloWorld)
	r.Patch("/patch", helloWorld)
	r.Delete("/delete", helloWorld)
	r.Route("/route", func(r chi.Router) {
		r.Get("/get", helloWorld)
		r.Post("/post", helloWorld)
		r.Put("/put", helloWorld)
		r.Patch("/patch", helloWorld)
		r.Delete("/delete", helloWorld)
		r.Route("/test", func(r chi.Router) {
			r.Get("/hello/{id}", helloWorld)
		})
		r.Mount("/test2", chi.NewRouter().Group(func(r chi.Router) {
			r.Get("/inner", helloWorld)
		}))
	})

	log.Print(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}

func TestSelfDescribeAPI(t *testing.T) {
	go runServer(8080)

	url, err := url.Parse("http://localhost:8080/route/test2")
	assert.NoError(t, err)

	v, err := http.DefaultClient.Do(&http.Request{Method: "OPTIONS", URL: url})
	assert.NoError(t, err)

	defer v.Body.Close()
	body, err := ioutil.ReadAll(v.Body)
	assert.NoError(t, err)

	assert.Empty(t, string(body))
}
