package server

import (
	"awesome-portal/backend/model"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func api_status(res http.ResponseWriter, req *http.Request) {
	if req.URL.Query().Has("test") {
		res.Write([]byte(fmt.Sprintf("yeah, there was a query: %s\n", req.URL.Query().Get("test"))))
	}

	if req.Body != nil {
		res.Write([]byte("there was a body !\n"))
		if body, err := io.ReadAll(req.Body); err == nil {
			res.Write([]byte(strings.ToUpper(string(body))))
		}

		res.Write([]byte("finished !"))
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Hello World 2 !"))
}

func Start_server() {
	server := http.NewServeMux()
	server.HandleFunc("/status", api_status)

	fmt.Println("Listening on:", model.AW_ROOT)
	err := http.ListenAndServe(":8080", server)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
}
