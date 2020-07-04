package web

import (
	"net/http"
)

func StartServer() *http.Server {

}

// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, r.URL.Path[1:])
// })
