package datasource

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cafebazaar/blacksmith/logging"
	"golang.org/x/crypto/scrypt"
)

type User struct {
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

func userFromNodeValue(value string) (*User, error) {
	var u User
	err := json.Unmarshal([]byte(value), &u)
	return &u, err
}

func (u *User) EmailAddress() string {
	return u.Email
}

func (u *User) InboxAddress() string {
	return u.InboxAddr
}

func (u *User) UID() string {
	return u.UIDStr
}

func (u *User) Info() map[string]interface{} {
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

func (u *User) UpdateInfo(values map[string]string) error {
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

func (u *User) HasPassword() bool {
	return u.Password != ""
}

func (u *User) encodePassword(plainPassword string, salt []byte) ([]byte, error) {
	password, err := scrypt.Key([]byte(plainPassword), salt, 16384, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	return password, nil
}

func (u *User) AcceptsPassword(plainPassword string, salt []byte) bool {
	encodedInputPassword, err := u.encodePassword(plainPassword, salt)
	if err != nil {
		logging.Debug(debugTag, "Error while encodePassword: %s", err)
		return false
	}

	userPassword, err := base64.StdEncoding.DecodeString(u.Password)
	if err != nil {
		return false
	}
	if userPassword == nil || encodedInputPassword == nil {
		return false
	}
	if len(userPassword) != len(encodedInputPassword) {
		return false
	}
	for i := range userPassword {
		if userPassword[i] != encodedInputPassword[i] {
			return false
		}
	}
	return true
}

func (u *User) SetPassword(plainPassword string, salt []byte) error {
	encodedPassword, err := u.encodePassword(plainPassword, salt)
	if err != nil {
		return err
	}
	u.Password = base64.StdEncoding.EncodeToString(encodedPassword)
	return nil
}

func (u *User) IsActive() bool {
	return u.Active
}

func (u *User) SetActive(active bool) {
	u.Active = active
}

func (u *User) IsAdmin() bool {
	return u.Admin
}

func (u *User) SetAdmin(admin bool) {
	u.Admin = admin
}
