package model

type AwesomeLink struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       uint   `json:"level"`
	UpdateTs    int64  `json:"updated"`
	OriginUrl   string `json:"originUrl"`
	ReadmeUrl   string `json:"-"`
	CloneUrl    string `json:"-"`
	OriginHash  string `json:"-"`
	Watchers    int    `json:"subscribersCount"`
	Subscribers int    `json:"watchersCount"`
	Topics      string `json:"-"`
}

type GHResponse struct {
	name      string `json:"string"`
	full_name string `json:"string"`
}

// link and it's metadata found while scanning an Awesome Index page (readme).
// it can become an AwesomeLink at some point
type PotentialLink struct {
	Parents     string
	Name        string
	Description string
	URL         string
}

// // act as a constructor for struct
// func NewPotentialLink() PotentialLink {
// 	self := PotentialLink{}
// 	self.parents = []string{}
// 	self.name = ""
// 	self.url = ""
// 	return self
// }
