package main

import (
	"fmt"
	"math/rand"
)

type operator struct {
	sign string
	fn   func(a, b int64) int64
}

var operators = []operator{
	{
		"+", func(a, b int64) int64 { return a + b },
	},
	{
		"-", func(a, b int64) int64 { return a - b },
	},
	{
		"*", func(a, b int64) int64 { return a * b },
	},
}

func generateProblem() (expression string, expectedAnswer int64) {
	a, b := rand.Int63n(15), rand.Int63n(15)
	op := operators[rand.Intn(len(operators))]

	return fmt.Sprintf("%d %s %d = ?", a, op.sign, b), op.fn(a, b)
}
