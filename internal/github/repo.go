package github

type SearchResult struct {
	TotalCount        int
	IncompleteResults bool
	Items             []Repo
}

type Repo struct {
	Id         int    `json:"id"`
	CreatedAt  string `json:"created_at"`
	FullName   string `json:"full_name"`
	Owner      `json:"owner"`
	Repository string `json:"name"`
	Languages  map[string]Language
}

type Owner struct {
	Owner string `json:"login"`
}

type Language struct {
	Bytes int
}

type RepoOut struct {
	FullName string
	Owner
	Repository string
	Languages  map[string]Language
}

// func (r *Repo) MarshalJSON() ([]byte, error) {
// 	ro := RepoOut(*r)
// 	return json.Marshal(&ro)
// }
