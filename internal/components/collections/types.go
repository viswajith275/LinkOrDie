package collections

import (
	collectionsdomain "link_or_die/internal/domains/collections"
	linkdomain "link_or_die/internal/domains/links"
	"strconv"
)

type PageProps struct {
	UserName    string
	Collections []collectionsdomain.Summary
}

type DetailProps struct {
	UserName   string
	Collection collectionsdomain.Collection
	Links      []linkdomain.View
}

type FormProps struct {
	Name  string
	Error string
}

func pluralize(n int, singular string) string {
	if n == 1 {
		return "1 " + singular
	}
	return strconv.Itoa(n) + " " + singular + "s"
}
