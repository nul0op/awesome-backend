package server

import (
	controller "awesome-portal/backend/controller"
	"awesome-portal/backend/model"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

// FIXME: put those globals somewhere else ?
var (
	firebaseAuth *auth.Client
	ctx          context.Context
	// app          *firebase.App
)

const (
	firebaseConfigFile = "path/to/your/firebaseConfig.json"
	firebaseDBURL      = "https://your-firebase-project.firebaseio.com"
)

func initAuth() {
	ctx = context.Background()
	opt := option.WithCredentialsJSON([]byte(os.Getenv("FIREBASE_CONFIG")))

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		fmt.Printf("Firebase error: cannot get app: %v\n", err)
	}

	firebaseAuth, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("Firebase error: cannot get auth client: %v\n", err)
	}
}

func apiStatus(res http.ResponseWriter, req *http.Request) {
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

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Hello World 2 !"))
}

// func appMiddleware(next http.Handler) http.handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Print("Executing appMiddleware")
// 		apiStatus()
// 		next.ServeHTTP(w, r)
// 		fmt.Print("Executing appMiddleware again")
// 	}
// }

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			fmt.Printf("got option: replying with cors allow all header\n")
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.Header().Add("Access-Control-Allow-Methods", "GET, OPTIONS, POST, DELETE, PUT")
			w.WriteHeader(http.StatusOK)
			return
		}

		fmt.Printf("\nMethod: %s\n", r.Method)
		fmt.Print("checking auth\n")

		for name, values := range r.Header {
			// Loop over all values for the name.
			for _, value := range values {
				fmt.Println(name, value)
			}
		}
		if len(r.Header.Get("Authorization")) == 0 {
			fmt.Print(("No authentication header found ! exiting\n"))
			w.Write([]byte("GFY !"))
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// we need to remove "Bearer: " prefix
		result, err := checkToken(r.Header.Get("Authorization")[7:])

		if err != nil {
			fmt.Printf("error while checking token: %v\n", err)
		} else {
			fmt.Print("auth return: ", result)
		}

		next.ServeHTTP(w, r)

	})
}

func middlewareOne(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("Executing middlewareOne\n")
		next.ServeHTTP(w, r)
		log.Print("Executing middlewareOne again\n")
	})
}

func finalMiddleware(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Final handler\n")
	w.Write([]byte("OK"))
}

func checkToken(token string) (uid string, err error) {
	result, err := firebaseAuth.VerifyIDToken(ctx, token)
	if err != nil {
		fmt.Printf("Firebase error: cannot verify token: %v\n", err)
	}
	fmt.Printf("Firebase success: token is: %v\n", result)
	return
}

func StartServer() {
	server := http.NewServeMux()
	initAuth()

	// finalHandler := http.HandlerFunc(finalMiddleware)
	server.Handle("/api/status", authMiddleware(http.HandlerFunc(apiStatus)))
	server.Handle("/links", authMiddleware(http.HandlerFunc(controller.GetLinks)))

	// server.HandleFunc("/api/status", apiStatus)

	fmt.Println("Listening on:", model.AW_ROOT)
	err := http.ListenAndServe(":3000", server)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
}
