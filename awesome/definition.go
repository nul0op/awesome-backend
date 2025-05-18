package definition

type AwesomeLink struct {
	// Name()	string
	// SetName( name string)

	name        string
	description string
	level       uint
	update_ts   uint64
	origin_url  string // first URL found that defines this link
	readme_url  string
	clone_url   string
	origin_hash string //sha256 hash of origin, used to "garante" unicity inside the index
	watchers    int
	subscribers int
	topics      []string
}

type GHResponse struct {
	name      string `json:"string"`
	full_name string `json:"string"`
}
