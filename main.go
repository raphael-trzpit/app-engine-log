package main

import (
	"cloud.google.com/go/logging"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"regexp"

	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
)


func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	ctx := context.Background()
	client, err := logging.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer client.Close()

	logger := client.Logger("my-log", logging.CommonResource(&mrpb.MonitoredResource{
		Type: "gae_app",
		Labels: map[string]string{
			"module_id": os.Getenv("GAE_SERVICE"),
			"project_id":  os.Getenv("GOOGLE_CLOUD_PROJECT"),
			"version_id": os.Getenv("GAE_VERSION"),
		},
	}))

	logger.Log(logging.Entry{Payload: "my test log"})
	http.Handle("/", indexHandler(logger))

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

// indexHandler responds to requests with our greeting.
func indexHandler(logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		reg := regexp.MustCompile(`^([a-z0-9]+)/.*`)
		logger.Log(logging.Entry{
			Payload: map[string]interface{}{ "message": "request", "context": "context"},
			HTTPRequest: &logging.HTTPRequest{Request: r},
			Trace: fmt.Sprintf("projects/%s/traces/%s", os.Getenv("GOOGLE_CLOUD_PROJECT"), reg.FindStringSubmatch(r.Header.Get("X-Cloud-Trace-Context"))[1]),
		})
		fmt.Fprint(w, "Hello, World!")
	})
}
