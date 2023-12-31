package github

import "encoding/json"

type Repo struct {
	Id           int    `json:"id"`
	FullName     string `json:"full_name"`
	Owner        `json:"owner"`
	Repository   string `json:"name"`
	LanguagesURL string `json:"languages_url"`
	Languages    map[string]Language
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

type SearchResult struct {
	TotalCount int    `json:"total_count"`
	Items      []Repo `json:"items"`
}

func (r *Repo) MarshalJSON() ([]byte, error) {
	ro := RepoOut{
		r.FullName,
		r.Owner,
		r.Repository,
		r.Languages,
	}
	return json.Marshal(&ro)
}
