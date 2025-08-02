package utils

import (
	"fmt"
	"strings"
)

func SizeAdjust(size string) (string,error ){

	size = strings.ToLower(size)
	size = strings.TrimSpace(size)
	size = strings.ReplaceAll(size, " ", "")

	switch size{
	case "large","l":
		return "L",nil
	case "medium","m":
		return "M",nil
	case "small","s":
		return "S",nil
	case "extrasmall","xs":
		return "XS",nil
	case "extralarge","xl":
		return "XL",nil
	case "xxlarge","xxl":
		return "XXL",nil
	case "uk 1","uk-1":
		return "UK-1",nil
	case "uk 2","uk-2":
		return "UK-2",nil
	case "uk 3":
		return "UK-3",nil
	case "uk 4","uk-4":
		return "UK-4",nil
	case "uk 5","uk-5":
		return "UK-5",nil
	case "uk 6","uk-6":
		return "UK-6",nil
	case "uk 7","uk-7":
		return "UK-7",nil
	case "uk 8","uk-8":
		return "UK-8",nil
	case "uk 9","uk-9":
		return "UK-9",nil
	case "uk 10","uk-10":
		return "UK-10",nil
	}

	return "",fmt.Errorf("invalid size: please use one of the predefined sizes")
}