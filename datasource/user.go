package datasource

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cafebazaar/blacksmith/logging"
	"golang.org/x/crypto/scrypt"
)

type User struct {
	Email         string `json:"email"`
	UIDStr        string `json:"uid"`
	InboxAddr     string `json:"inboxAddress"`
	Active        bool   `json:"active,bool,omitempty"`
	Admin         bool   `json:"admin,bool,omitempty"`
	Password      string `json:"password,omitempty"`
	EnFirstName   string `json:"enFirstName,omitempty"`
	EnLastName    string `json:"enLastName,omitempty"`
	FaFirstName   string `json:"faFirstName,omitempty"`
	FaLastName    string `json:"faLastName,omitempty"`
	MobileNum     string `json:"mobileNum,omitempty"`
	EmergencyNum  string `json:"emergencyNum,omitempty"`
	BirthDate     uint64 `json:"birthDate,omitempty"`
	EnrolmentDate uint64 `json:"enrolmentDate,omitempty"`
	LeavingDate   uint64 `json:"leavingDate,omitempty"`
	// Links         []string `json:"birthDate,array,omitempty"`
}

func userFromNodeValue(value string) (*User, error) {
	var u User
	err := json.Unmarshal([]byte(value), &u)
	return &u, err
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
