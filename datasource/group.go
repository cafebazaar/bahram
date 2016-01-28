package datasource

import (
	"encoding/json"
)

type Group struct {
	Email       string   `json:"email,string"`
	Name        string   `json:"name,string"`
	Description string   `json:"description,string"`
	Active      bool     `json:"active,bool"`
	Joinable    bool     `json:"joinable,bool"`
	Manager     string   `json:"manager,string"`
	MembersLst  []string `json:"members,array"`
	CCs         []string `json:"ccs,array"`
}

func groupFromNodeValue(value string) (*Group, error) {
	var g Group
	err := json.Unmarshal([]byte(value), &g)
	return &g, err
}

func (g *Group) EmailAddress() string {
	return g.Email
}

func (g *Group) MembersList() []string {
	return g.MembersLst
}
