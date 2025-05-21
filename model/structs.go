package model

type AwesomeLink struct {
	// Name()	string
	// SetName( name string)

	Name        string
	Description string
	Level       uint
	UpdateTs    uint64
	OriginUrl   string // first URL found that defines this link
	ReadmeUrl   string
	CloneUrl    string
	OriginHash  string //sha256 hash of origin, used to "garante" unicity inside the index
	Watchers    int
	Subscribers int
	Topics      []string
}

type GHResponse struct {
	name      string `json:"string"`
	full_name string `json:"string"`
}
