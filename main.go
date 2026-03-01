package main

import (
	"log"
	"net/http"
	"strconv"
	"text/template"

	"pr6/calc"
)

type pageData struct {
	Variant int
	Result  *calc.Result
	Error   string
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	http.HandleFunc("/", handleIndex)

	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := pageData{
		Variant: 1,
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			data.Error = "Не вдалося розібрати форму: " + err.Error()
			render(w, data)
			return
		}

		variantStr := r.FormValue("variant")
		v, err := strconv.Atoi(variantStr)
		if err != nil {
			data.Error = "Варіант має бути цілим числом від 0 до 9"
			render(w, data)
			return
		}

		data.Variant = v

		res, err := calc.Calculate(v)
		if err != nil {
			data.Error = err.Error()
		} else {
			data.Result = res
		}
	}

	render(w, data)
}

func render(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Помилка шаблону: "+err.Error(), http.StatusInternalServerError)
	}
}

