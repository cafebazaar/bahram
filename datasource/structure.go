package datasource // import "github.com/cafebazaar/bahram/datasource"

const (
	debugTag = "DATASOURCE"
)

type User interface {
	InboxAddress() string
	UID() string
	Info() map[string]string
}

type Group interface {
	Users() ([]User, error)
}

type DataSource interface {
	CreateUser(active bool, values map[string]string) (User, error)
	UserByEmail(emailAddress string) (User, error)
	GroupByEmail(emailAddress string) (Group, error)
}
