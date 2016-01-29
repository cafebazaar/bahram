package datasource

import (
	"encoding/json"
)

type Group struct {
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Active      bool     `json:"active"`
	Public      bool     `json:"public"`
	Manager     string   `json:"manager"`
	Members     []string `json:"members"`
	CCs         []string `json:"ccs"`
}

func groupFromNodeValue(value string) (*Group, error) {
	var g Group
	err := json.Unmarshal([]byte(value), &g)
	return &g, err
}

func (g *Group) IsMemeber(email string) bool {
	if g.Manager == email {
		return true
	}
	for _, m := range g.Members {
		if m == email {
			return true
		}
	}
	return false
}
