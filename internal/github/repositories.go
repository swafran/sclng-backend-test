package github

import "encoding/json"

type Repos []Repo

type Repo struct {
	Repository string
	FullName   string
	Owner      string
}

func (r *Repo) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		return nil
	}

	var ghRepo struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	}

	if err := json.Unmarshal(b, &ghRepo); err != nil {
		return err
	}

	*r = Repo{
		Repository: ghRepo.Name,
		FullName:   ghRepo.FullName,
		Owner:      ghRepo.Owner.Login,
	}
}
