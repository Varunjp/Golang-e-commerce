package utils

func GetPaginationPages(currentPage, totalPages int) []int {
	maxPagesToShow := 25
	var pages []int

	start := currentPage - maxPagesToShow/2
	if start < 1 {
		start = 1
	}
	end := start + maxPagesToShow - 1
	if end > totalPages {
		end = totalPages
	}
	// adjust start again if total pages less than maxPagesToShow
	if end-start+1 < maxPagesToShow {
		start = end - maxPagesToShow + 1
		if start < 1 {
			start = 1
		}
	}

	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}