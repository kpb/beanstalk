package beans

import "strings"

// Search returns beans whose ID, title, or Markdown body contains the query.
// The caller controls whether archived beans are included and the result order.
func Search(loaded []Bean, query string) []Bean {
	query = strings.ToLower(strings.TrimSpace(query))
	matches := make([]Bean, 0)
	if query == "" {
		return matches
	}
	for _, bean := range loaded {
		if strings.Contains(strings.ToLower(bean.ID), query) ||
			strings.Contains(strings.ToLower(bean.Title), query) ||
			strings.Contains(strings.ToLower(bean.Body), query) {
			matches = append(matches, bean)
		}
	}
	return matches
}
