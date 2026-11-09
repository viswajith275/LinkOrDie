package layout

import "github.com/a-h/templ"

type AppProps struct {
	Title    string
	UserName string
	Active   string
	Content  templ.Component
}

type AuthProps struct {
	Title   string
	Content templ.Component
}

type BareProps struct {
	Title   string
	Content templ.Component
}
