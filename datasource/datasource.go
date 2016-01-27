package datasource // import "github.com/cafebazaar/bahram/datasource"

import (
	etcd "github.com/coreos/etcd/client"
)

type User interface {
	Groups() []Group
}

type Group interface {
	Users() []User
}

type DataSource interface {
	UserByID(id string) User
	UserByEmail(emailAddress string) User
	GroupByEmail(emailAddress string) Group
}

func NewEtcdDataSource(kapi etcd.KeysAPI, client etcd.Client) (DataSource, error) {
	return nil, nil
}
