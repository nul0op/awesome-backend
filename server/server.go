package server

import (
	controller "awesome-portal/backend/controller"
	"awesome-portal/backend/model"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var (
	firebaseAuth *auth.Client
	ctx          context.Context
)

func initAuth() {
	ctx = context.Background()
	opt := option.WithCredentialsJSON([]byte(os.Getenv("FIREBASE_CONFIG")))

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		model.Log.Error("Firebase: cannot get app:", "error", err)
	}

	firebaseAuth, err = app.Auth(ctx)
	if err != nil {
		model.Log.Error("Firebase: cannot get auth client:", "error", err)
	}
}

// random things to test request handling ...
func apiStatus(res http.ResponseWriter, req *http.Request) {
	if req.URL.Query().Has("test") {
		res.Write([]byte(fmt.Sprintf("there was a query: %s\n", req.URL.Query().Get("test"))))
	}

	if req.Body != nil {
		res.Write([]byte("there was a body: converting it in Uppercase\n"))
		if body, err := io.ReadAll(req.Body); err == nil {
			res.Write([]byte(strings.ToUpper(string(body))))
		}

		res.Write([]byte("finished !"))
	}

	res.WriteHeader(http.StatusOK)
}

func addCORSHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type,Authorization")
	w.Header().Add("Access-Control-Allow-Methods", "GET, OPTIONS")
}

// validate Firebase token, set CORS
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		addCORSHeaders(w, r)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// debug
		// for name, values := range r.Header {
		// 	for _, value := range values {
		// 		fmt.Println(name, value)
		// 	}
		// }

		if len(r.Header.Get("Authorization")) == 0 {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("No authentication header found, get out !"))
			return
		}

		// validate Firebase token (need to remove "Bearer: " prefix)
		uid, err := checkToken(r.Header.Get("Authorization")[7:])
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			model.Log.Warnf("unable to check token: %v\n", err)
			return

		} else {
			model.Log.Debug("uid [%s] successfuly authenticated", uid)
		}

		next.ServeHTTP(w, r)
	})
}

func checkToken(token string) (uid string, err error) {
	uid = ""

	result, err := firebaseAuth.VerifyIDToken(ctx, token)
	if err != nil {
		model.Log.Errorf("Firebase: unable to verify token: %v", err)
		return
	}

	model.Log.Debugf("Firebase success: auth package is: [%v]", result)
	uid = result.UID
	return
}

func StartServer() {

	server := http.NewServeMux()
	initAuth()

	server.Handle("/api/status", authMiddleware(http.HandlerFunc(apiStatus)))
	server.Handle("/links", authMiddleware(http.HandlerFunc(controller.GetLinks)))

	// main.ALogger.Info("Listening on [%s]\n", os.Getenv("API_PORT"))
	err := http.ListenAndServe(":"+os.Getenv("API_PORT"), server)
	if err != nil {
		model.Log.Error(err)
	}
}
