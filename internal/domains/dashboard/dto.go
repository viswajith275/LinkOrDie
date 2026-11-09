package dashboard

import (
	"link_or_die/internal/domains/collections"
	"link_or_die/internal/domains/links"
)

type Stats struct {
	Total       int
	Alive       int
	Flaky       int
	Dead        int
	Pending     int
	Collections int
}

type DashboardView struct {
	UserName    string
	Email       string
	Stats       Stats
	Collections []collections.Summary
	RecentLinks []links.View
}
