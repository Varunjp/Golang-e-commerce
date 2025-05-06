package utils

import "html/template"

// TemplateFuncs returns reusable helper functions for templates
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"iterate": func(start, end int) []int {
			s := make([]int, end-start+1)
			for i := range s {
				s[i] = start + i
			}
			return s
		},
		"slice": slice,
		"add1": add1,

	}
}

func slice(vals ...int)[]int{
	return vals
}

func add1(i int) int {
    return i + 1
}