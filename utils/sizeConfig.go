package utils

import "strings"

func SizeAdjust(size string) string {

	size = strings.ToLower(size)

	switch size{
	case "large":
		return "L"
	case "medium":
		return "M"
	case "small":
		return "S"
	case "extrasmall":
		return "XS"
	}

	return strings.ToUpper(size)
}