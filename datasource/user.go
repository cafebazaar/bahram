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
	Active        bool   `json:"active"`
	Admin         bool   `json:"admin"`
	Password      string `json:"password"`
	EnFirstName   string `json:"enFirstName"`
	EnLastName    string `json:"enLastName"`
	FaFirstName   string `json:"faFirstName"`
	FaLastName    string `json:"faLastName"`
	MobileNum     string `json:"mobileNum"`
	EmergencyNum  string `json:"emergencyNum"`
	BirthDate     uint64 `json:"birthDate"`
	EnrolmentDate uint64 `json:"enrolmentDate"`
	LeavingDate   uint64 `json:"leavingDate"`
	// Links         []string `json:"birthDate"`
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
	// TODO can we encode random salt into encoded password
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
