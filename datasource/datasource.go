package datasource

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cafebazaar/blacksmith/logging"
	etcd "github.com/coreos/etcd/client"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
)

type EtcdDataSource struct {
	keysAPI etcd.KeysAPI
	etcdDir string
	cache   *cache.Cache
}

func NewEtcdDataSource(kapi etcd.KeysAPI, etcdDir string) (DataSource, error) {
	instance := &EtcdDataSource{
		keysAPI: kapi,
		etcdDir: etcdDir,
		cache:   cache.New(1*time.Minute, 30*time.Second), // protects against brute force
	}

	return instance, nil
}

func (ds *EtcdDataSource) ConfigString(name string) string {
	return os.Getenv(fmt.Sprintf("BAHRAM_%s", name))
}

func (ds *EtcdDataSource) ConfigByteArray(name string) []byte {
	base64Value := ds.ConfigString(name)
	value, err := base64.StdEncoding.DecodeString(base64Value)
	if err != nil {
		logging.Log(debugTag, "Error while decoding config value %s: %s", name, err)
		return nil
	}
	return value
}

func (ds *EtcdDataSource) CreateUser(emailAddress, uid, inboxAddress string) (User, error) {

	if emailAddress == "" || uid == "" || inboxAddress == "" {
		return nil, errors.New("emailAddress, uid, and inboxAddress is required.")
	}

	_, err := ds.UserByEmail(emailAddress)
	if err == nil {
		return nil, errors.New("A user with this email already exists")
	}

	_, err = ds.GroupByEmail(emailAddress)
	if err == nil {
		return nil, errors.New("A group with this email already exists")
	}

	return user(ds, emailAddress, uid, inboxAddress)
}

func (ds *EtcdDataSource) StoreUser(u User) error {
	userJSON, err := json.Marshal(u)
	if err != nil {
		return err
	}

	logging.Debug(debugTag, "Setting %s", userJSON)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = ds.keysAPI.Set(ctx, fmt.Sprintf("/%s/users/%s", ds.etcdDir, u.EmailAddress()), string(userJSON[:]), nil)
	if err != nil {
		return err
	}
	return nil
}

func (ds *EtcdDataSource) UserByEmail(emailAddress string) (User, error) {
	// TODO: use ds.cache

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response, err := ds.keysAPI.Get(ctx, fmt.Sprintf("/%s/users/%s", ds.etcdDir, emailAddress), nil)
	if err != nil {
		return nil, err
	}

	return userFromNodeValue(ds, response.Node.Value)
}

func (ds *EtcdDataSource) StoreGroup(g *Group) error {
	groupJSON, err := json.Marshal(g)
	if err != nil {
		return err
	}

	logging.Debug(debugTag, "Setting %s", groupJSON)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = ds.keysAPI.Set(ctx, fmt.Sprintf("/%s/groups/%s", ds.etcdDir, g.EmailAddress()), string(groupJSON[:]), nil)
	if err != nil {
		return err
	}
	return nil
}

func (ds *EtcdDataSource) GroupByEmail(emailAddress string) (*Group, error) {
	// TODO: use ds.cache

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response, err := ds.keysAPI.Get(ctx, fmt.Sprintf("/%s/groups/%s", ds.etcdDir, emailAddress), nil)
	if err != nil {
		return nil, err
	}

	return groupFromNodeValue(response.Node.Value)
}

func (ds *EtcdDataSource) Groups() ([]Group, error) {
	return nil, nil
}
