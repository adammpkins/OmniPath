package docs

import (
	"log"
	"net/http"
)

// ServeLocalDocs serves documentation from the specified directory on the given port.
func ServeLocalDocs(dir string, port string) {
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	log.Printf("Serving local documentation from %s at http://localhost:%s\n", dir, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
