package datasource // import "github.com/cafebazaar/bahram/datasource"

type User interface {
	Groups() ([]Group, error)
	InboxAddress() string
}

type Group interface {
	Users() ([]User, error)
}

type DataSource interface {
	UserByID(id string) (User, error)
	UserByEmail(emailAddress string) (User, error)
	GroupByEmail(emailAddress string) (Group, error)
}
