package controller

import (
	"awesome-portal/backend/model"
	"encoding/json"
	"net/http"
)

func GetLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	result := model.GetLinks()
	json, _ := json.Marshal(result)

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
