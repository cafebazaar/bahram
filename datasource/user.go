package datasource // import "github.com/cafebazaar/bahram/datasource"

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type userImpl struct {
	Email         string `json:"email,string"`
	UIDStr        string `json:"uid,string"`
	InboxAddr     string `json:"inboxAddress,string"`
	Active        bool   `json:"active,bool"`
	Admin         bool   `json:"admin,bool"`
	Password      string `json:"password,string"`
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

func (u *userImpl) EmailAddress() string {
	return u.Email
}

func (u *userImpl) InboxAddress() string {
	return u.InboxAddr
}

func (u *userImpl) UID() string {
	return u.UIDStr
}

func (u *userImpl) Info() map[string]interface{} {
	return map[string]interface{}{
		"email":         u.Email,
		"uid":           u.UIDStr,
		"enFirstName":   u.EnFirstName,
		"enLastName":    u.EnLastName,
		"faFirstName":   u.FaFirstName,
		"faLastName":    u.FaLastName,
		"mobileNum":     u.MobileNum,
		"emergencyNum":  u.EmergencyNum,
		"birthDate":     u.BirthDate,
		"enrolmentDate": u.EnrolmentDate,
		"leavingDate":   u.LeavingDate,
	}
}

func (u *userImpl) UpdateInfo(values map[string]string) error {
	enFirstName, ok := values["enFirstName"]
	if ok {
		u.EnFirstName = enFirstName
	}

	enLastName, ok := values["enLastName"]
	if ok {
		u.EnLastName = enLastName
	}

	faFirstName, ok := values["faFirstName"]
	if ok {
		u.FaFirstName = faFirstName
	}

	faLastName, ok := values["faLastName"]
	if ok {
		u.FaLastName = faLastName
	}

	mobileNum, ok := values["mobileNum"]
	if ok {
		u.MobileNum = mobileNum
	}

	emergencyNum, ok := values["emergencyNum"]
	if ok {
		u.EmergencyNum = emergencyNum
	}

	var err error

	birthDateStr, ok := values["birthDate"]
	if ok {
		u.BirthDate, err = strconv.ParseUint(birthDateStr, 10, 64)
		if err != nil {
			return fmt.Errorf("error while parsing birthDate: %s", err)
		}
	}

	enrolmentDateStr, ok := values["enrolmentDate"]
	if ok {
		u.EnrolmentDate, err = strconv.ParseUint(enrolmentDateStr, 10, 64)
		if err != nil {
			return fmt.Errorf("error while parsing enrolmentDate: %s", err)
		}
	}

	leavingDateStr, ok := values["leavingDate"]
	if ok {
		u.LeavingDate, err = strconv.ParseUint(leavingDateStr, 10, 64)
		if err != nil {
			return fmt.Errorf("error while parsing leavingDate: %s", err)
		}
	}

	return nil
}

func (u *userImpl) AcceptsPassword(plainPassword string) bool {
	return true
}

func (u *userImpl) SetPassword(plainPassword string) {

}

func (u *userImpl) IsActive() bool {
	return u.Active
}

func (u *userImpl) SetActive(active bool) {
	u.Active = active
}

func (u *userImpl) IsAdmin() bool {
	return u.Admin
}

func (u *userImpl) SetAdmin(admin bool) {
	u.Admin = admin
}
