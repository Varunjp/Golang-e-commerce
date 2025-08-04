package utils

import (
	"html/template"
	"math"
	"time"
)

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
		"formatDate":formatDate,
		"round2":round2,
		"seq":sec,
		"floor":floor,
		"toFloat":tofloat,
		"inSlice":inSlice,
		"until":until,
		"max":Max,
		"min":Min,
		"seqs":Secq,
		"toInt": func(u uint) int {
        	return int(u)
    	},
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

func formatDate(t time.Time)string{
	return t.Format("01-02-2006")
}

func round2(f float64)float64{
	return math.Round(f*100)/100
}

func sec(n int) []int{
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

func floor (x float64) float64{
	return math.Floor(x)
}

func tofloat(i int)float64{
	return float64(i)
}

func inSlice(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func until(n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = i
	}
	return result
}

func Max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func Min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func Secq(start, end int) []int {
    s := make([]int, end-start+1)
    for i := range s {
        s[i] = start + i
    }
    return s
}