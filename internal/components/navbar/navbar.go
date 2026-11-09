package navbar

type Props struct {
	UserName string
	Active   string
}

func navClass(active, name string) string {
	if active == name {
		return "nav-link active"
	}
	return "nav-link"
}
