package datasource // import "github.com/cafebazaar/bahram/datasource"

const (
	debugTag = "DATASOURCE"
)

type User interface {
	EmailAddress() string
	InboxAddress() string
	UID() string
	Info() map[string]interface{}
	UpdateInfo(values map[string]string) error
	HasPassword() bool
	AcceptsPassword(plainPassword string) bool
	SetPassword(plainPassword string) error
	IsActive() bool
	SetActive(active bool)
	IsAdmin() bool
	SetAdmin(admin bool)
}

type Group interface {
	Users() ([]User, error)
}

type DataSource interface {
	CreateUser(emailAddress, uid, inboxAddress string) (User, error)
	StoreUser(u User) error
	UserByEmail(emailAddress string) (User, error)
	GroupByEmail(emailAddress string) (Group, error)
	ConfigString(name string) string
	ConfigByteArray(name string) []byte
}
