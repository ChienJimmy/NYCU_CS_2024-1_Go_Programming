package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// TODO: Create a struct to hold the data sent to the template
type Calculation struct {
	Expression string
	Result     string
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func lcm(a, b int) int {
	return a * b / gcd(a, b)
}

func Calculator(w http.ResponseWriter, r *http.Request) {
	// TODO: Finish this function
	op := r.URL.Query().Get("op")
	num1Str := r.URL.Query().Get("num1")
	num2Str := r.URL.Query().Get("num2")

	num1, err1 := strconv.Atoi(num1Str)
	num2, err2 := strconv.Atoi(num2Str)

	if err1 != nil || err2 != nil || num2 == 0 && op == "div" {
		http.ServeFile(w, r, "error.html")
		return
	}

	var result int64
	var expression string

	switch op {
	case "add":
		result = int64(num1 + num2)
		expression = fmt.Sprintf("%d + %d", num1, num2)
	case "sub":
		result = int64(num1 - num2)
		expression = fmt.Sprintf("%d - %d", num1, num2)
	case "mul":
		result = int64(num1 * num2)
		expression = fmt.Sprintf("%d * %d", num1, num2)
	case "div":
		result = int64(float64(num1) / float64(num2))
		expression = fmt.Sprintf("%d / %d", num1, num2)
	case "gcd":
		result = int64(gcd(num1, num2))
		expression = fmt.Sprintf("GCD(%d, %d)", num1, num2)
	case "lcm":
		result = int64(lcm(num1, num2))
		expression = fmt.Sprintf("LCM(%d, %d)", num1, num2)
	default:
		http.ServeFile(w, r, "error.html")
		return
	}

	calculation := Calculation{
		Expression: expression,
		Result:     fmt.Sprintf("%d", result),
	}
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.ServeFile(w, r, "error.html")
		return
	}
	err = tmpl.Execute(w, calculation)
	if err != nil {
		http.ServeFile(w, r, "error.html")
		return
	}
}

func main() {
	http.HandleFunc("/", Calculator)
	log.Fatal(http.ListenAndServe(":8084", nil))
}
