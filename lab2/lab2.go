package main

import (
	"fmt"
	"strings"
)

func main() {
	var n int64

	fmt.Print("Enter a number: ")
	fmt.Scanln(&n)

	result := Sum(n)
	fmt.Println(result)
}

func Sum(n int64) string {
	// TODO: Finish this function
	var numbers []string
	sum := int64(0)
	for i := int64(1); i <= n; i++ {
		if i%7 != 0 {
			sum += i
			numbers = append(numbers, fmt.Sprintf("%d", i))
		}
	}
	return fmt.Sprintf("%s=%d", strings.Join(numbers, "+"), sum)
}
