package main

import "net/http"
import "github.com/gorilla/handlers"
import "log"
import "os"

func main() {
	apiEnv := &ApiEnv{
		Namespaces: map[string]NamespaceConfig{
			"": NamespaceConfig{
				BasePath: ".\\www\\",
			},
		},
	}

	apiRouter := NewAPI(apiEnv)

	loggedRouter := handlers.LoggingHandler(os.Stdout, apiRouter)
	log.Fatal(http.ListenAndServe(":3000", loggedRouter))
}
