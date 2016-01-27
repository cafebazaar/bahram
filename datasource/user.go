package datasource // import "github.com/cafebazaar/bahram/datasource"

import (
	"encoding/json"
	"strconv"
)

type userImpl struct {
	Email         string `json:"email,string"`
	UIDStr        string `json:"uid,string"`
	InboxAddr     string `json:"inboxAddress,string"`
	Active        bool   `json:"active,bool"`
	EnFirstName   string `json:"enFirstName,string,omitempty"`
	EnLastName    string `json:"enLastName,string,omitempty"`
	FaFirstName   string `json:"faFirstName,string,omitempty"`
	FaLastName    string `json:"faLastName,string,omitempty"`
	MobileNum     string `json:"mobileNum,string,omitempty"`
	EmergencyNum  string `json:"emergencyNum,string,omitempty"`
	BirthDate     uint64 `json:"birthDate,string,omitempty"`
	EnrolmentDate uint64 `json:"enrolmentDate,string,omitempty"`
	LeavingDate   uint64 `json:"leavingDate,string,omitempty"`
	// Links         []string `json:"birthDate,array,omitempty"`
}

func userFromNodeValue(value string) (User, error) {
	var u userImpl
	err := json.Unmarshal([]byte(value), &u)
	return &u, err
}

func (u *userImpl) InboxAddress() string {
	return u.InboxAddr
}

func (u *userImpl) UID() string {
	return u.UIDStr
}

func (u *userImpl) Info() map[string]string {
	return map[string]string{
		"email":         u.Email,
		"uid":           u.UIDStr,
		"enFirstName":   u.EnFirstName,
		"enLastName":    u.EnLastName,
		"faFirstName":   u.FaFirstName,
		"faLastName":    u.FaLastName,
		"mobileNum":     u.MobileNum,
		"emergencyNum":  u.EmergencyNum,
		"birthDate":     strconv.FormatUint(u.BirthDate, 10),
		"enrolmentDate": strconv.FormatUint(u.EnrolmentDate, 10),
		"leavingDate":   strconv.FormatUint(u.LeavingDate, 10),
	}
}
