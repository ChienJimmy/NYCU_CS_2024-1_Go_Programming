package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func Calculator(w http.ResponseWriter, r *http.Request) {
	// TODO: implement a calculator
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		fmt.Fprintf(w, "Error!")
		return
	}
	operation := parts[1]
	num1, err1 := strconv.Atoi(parts[2])
	num2, err2 := strconv.Atoi(parts[3])
	if err1 != nil || err2 != nil {
		fmt.Fprintf(w, "Error!")
		return
	}
	switch operation {
	case "add":
		fmt.Fprintf(w, "%d + %d = %d", num1, num2, (num1 + num2))
	case "sub":
		fmt.Fprintf(w, "%d - %d = %d", num1, num2, (num1 - num2))
	case "mul":
		fmt.Fprintf(w, "%d * %d = %d", num1, num2, (num1 * num2))
	case "div":
		if num2 == 0 {
			fmt.Fprintf(w, "Error!")
			return
		}
		fmt.Fprintf(w, "%d / %d = %d, remainder = %d", num1, num2, (num1 / num2), (num1 % num2))
	default:
		fmt.Fprintf(w, "Error!")
		return
	}
}

func main() {
	http.HandleFunc("/", Calculator)
	log.Fatal(http.ListenAndServe(":8083", nil))
}
