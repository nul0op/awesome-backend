package model

type AwesomeLink struct {
	// Name()	string
	// SetName( name string)

	Name        string
	Description string
	Level       uint
	UpdateTs    int64
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

// link and it's metadata found while scanning an Awesome Index page (readme).
// it can become an AwesomeLink at some point
type PotentialLink struct {
	Parents []string
	Name    string
	Name2   string
	URL     string
}

// // act as a constructor for struct
// func NewPotentialLink() PotentialLink {
// 	self := PotentialLink{}
// 	self.parents = []string{}
// 	self.name = ""
// 	self.url = ""
// 	return self
// }
