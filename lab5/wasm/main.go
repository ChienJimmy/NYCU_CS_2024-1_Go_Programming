package main

import (
	"fmt"
	"math/big"
	"syscall/js"
)

func CheckPrime(this js.Value, args []js.Value) interface{} {
	// TODO: Check if the number is prime
	//
	// Retrieve input value from the HTML input element
	input := js.Global().Get("document").Call("getElementById", "value").Get("value").String()

	// Convert the input string to a big.Int
	num := new(big.Int)
	_, ok := num.SetString(input, 10)

	if !ok {
		js.Global().Get("document").Call("getElementById", "answer").Set("innerHTML", "Invalid input")
		return nil
	}

	// Check if the number is prime
	if num.ProbablyPrime(0) {
		js.Global().Get("document").Call("getElementById", "answer").Set("innerHTML", "It's prime")
	} else {
		js.Global().Get("document").Call("getElementById", "answer").Set("innerHTML", "It's not prime")
	}

	return nil
}

func registerCallbacks() {
	// TODO: Register the function CheckPrime
	js.Global().Set("CheckPrime", js.FuncOf(CheckPrime))
}

func main() {
	fmt.Println("Golang main function executed")
	registerCallbacks()

	//need block the main thread forever
	select {}
}
