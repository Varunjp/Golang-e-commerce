package utils

import "html/template"

// TemplateFuncs returns reusable helper functions for templates
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"gt": func(a,b int) bool { return a > b},
		"gtf": func(a,b float64) bool {return a > b},
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
		"iter":iter,
		"itere":itere,
		"mulFloat":mulFloat,
		"addFloat":addFloat,
	}
}

func slice(vals ...int)[]int{
	return vals
}

func add1(i int) int {
    return i + 1
}

func iter(count int) []int {
    var i []int
    for x := 0; x < count; x++ {
        i = append(i, x)
    }
    return i
}

func itere(count int) []int {
	var i int
	var items []int
	for i = 1; i <= count; i++ {
		items = append(items, i)
	}
	return items
}

func mulFloat(a float64, b int) float64 {
    return a * float64(b)
}

func addFloat(a,b float64) float64{
	return a + b
}