package controller

import (
	"fmt"
	"net/http"
)

func GetLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		fmt.Printf("got option: replying with cors allow all header\n")

		// w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		// w.Header().Add("Access-Control-Allow-Origin", "*")
		// w.Header().Add("Access-Control-Allow-Headers", "Content-Type,Authorization")
		// w.Header().Add("Access-Control-Allow-Methods", "GET, OPTIONS, POST, DELETE, PUT")
		w.WriteHeader(http.StatusOK)
		return
	}

	// if req.URL.Query().Has("test") {
	// 	res.Write([]byte(fmt.Sprintf("yeah, there was a query: %s\n", req.URL.Query().Get("test"))))
	// }

	// if req.Body != nil {
	// 	res.Write([]byte("there was a body !\n"))
	// 	if body, err := io.ReadAll(req.Body); err == nil {
	// 		res.Write([]byte(strings.ToUpper(string(body))))
	// 	}

	// 	res.Write([]byte("finished !"))
	// }

	fmt.Printf("got a query: ")

	// w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("{ link: 'brol' }"))
	// w.WriteHeader(http.StatusOK)
	fmt.Printf("reply sent\n")

}
