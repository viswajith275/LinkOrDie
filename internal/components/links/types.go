package links

import (
	collectionsdomain "link_or_die/internal/domains/collections"
	domain "link_or_die/internal/domains/links"

	"github.com/google/uuid"
)

type TableProps struct {
	Links          []domain.View
	ShowCollection bool
}

type PageProps struct {
	UserName string
	State    string
	Links    []domain.View
}

type DetailProps struct {
	UserName string
	Detail   domain.Detail
}

type FormProps struct {
	Collections          []collectionsdomain.Collection
	SelectedCollectionID *uuid.UUID
	URL                  string
	Title                string
	Error                string
}

func filterClass(active, value string) string {
	if active == value {
		return "active"
	}
	return ""
}
