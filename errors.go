package main

import (
	"forum/utils"
	"net/http"
)

func error404Page(w http.ResponseWriter, r *http.Request) {
	err := utils.Tmpl.ExecuteTemplate(w, "error404.html", nil)
	if err != nil {
		http.Error(w, "Error-Internal 500", http.StatusInternalServerError)
	}
}
