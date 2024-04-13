package certificate

import "net/http"

func check() {
	http.HandleFunc("/file/", func(w http.ResponseWriter, r *http.Request) {
		fileName := r.URL.Path[len("/file/"):]
		http.ServeFile(w, r, "/path/to/files/"+fileName)
	})
}

func serverChallengeFile() {
	http.HandleFunc("/.well-known/acme-challenge/{filename}", func(w http.ResponseWriter, r *http.Request) {

	})
}
