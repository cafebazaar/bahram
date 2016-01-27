package datasource

import (
	etcd "github.com/coreos/etcd/client"
)

type EtcdDataSource struct {
	keysAPI etcd.KeysAPI
}

func NewEtcdDataSource(kapi etcd.KeysAPI) (DataSource, error) {
	instance := &EtcdDataSource{
		keysAPI: kapi,
	}

	return instance, nil
}

func (ds *EtcdDataSource) UserByID(id string) (User, error) {
	return nil, nil
}

func (ds *EtcdDataSource) UserByEmail(emailAddress string) (User, error) {
	return nil, nil
}

func (ds *EtcdDataSource) GroupByEmail(emailAddress string) (Group, error) {
	return nil, nil
}
