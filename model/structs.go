package model

type AwesomeLink struct {
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Level       int    `json:"level" db:"level"`
	UpdateTs    int64  `json:"updated" db:"updated"`
	OriginUrl   string `json:"originUrl" db:"origin_url"`
	ReadmeUrl   string `json:"-" db:"-"`
	CloneUrl    string `json:"-" db:"-"`
	OriginHash  string `json:"-" db:"external_id"`
	Watchers    int    `json:"subscribersCount" db:"watchers_count"`
	Subscribers int    `json:"watchersCount" db:"subscribers_count"`
	Topics      string `json:"-" db:"topics"`
}

func NewAwesomeLink() AwesomeLink {
	al := AwesomeLink{}

	al.Name = ""
	al.Description = ""
	al.Level = 0
	al.UpdateTs = 0
	al.OriginUrl = ""
	al.ReadmeUrl = ""
	al.CloneUrl = ""
	al.OriginHash = ""
	al.Watchers = 0
	al.Subscribers = 0
	al.Topics = ""

	return al
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
