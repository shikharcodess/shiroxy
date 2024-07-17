package api

type Domains struct {
	Name   string
	Status string
	Token  string
}

type API struct {
	domains map[string]*Domains
}
