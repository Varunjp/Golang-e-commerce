package utils

import "strings"

func SizeAdjust(size string) string {

	size = strings.ToLower(size)
	size = strings.TrimSpace(size)
	size = strings.ReplaceAll(size, " ", "")

	switch size{
	case "large":
		return "L"
	case "medium":
		return "M"
	case "small":
		return "S"
	case "extrasmall":
		return "XS"
	case "extralarge":
		return "XL"
	case "uk 1":
		return "UK-1"
	case "uk 2":
		return "UK-2"
	case "uk 3":
		return "UK-3"
	case "uk 4":
		return "UK-4"
	case "uk 5":
		return "UK-5"
	case "uk 6":
		return "UK-6"
	case "uk 7":
		return "UK-7"
	case "uk 8":
		return "UK-8"
	case "uk 9":
		return "UK-9"
	case "uk 10":
		return "UK-10"
	}

	return strings.ToUpper(size)
}