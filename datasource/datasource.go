package datasource // import "github.com/cafebazaar/bahram/datasource"

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cafebazaar/blacksmith/logging"
	etcd "github.com/coreos/etcd/client"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
)

const (
	debugTag = "DATASOURCE"
)

type DataSource struct {
	keysAPI etcd.KeysAPI
	etcdDir string
	cache   *cache.Cache
}

func NewDataSource(kapi etcd.KeysAPI, etcdDir string) (*DataSource, error) {
	instance := &DataSource{
		keysAPI: kapi,
		etcdDir: etcdDir,
		cache:   cache.New(1*time.Minute, 30*time.Second), // protects against brute force
	}

	return instance, nil
}

func (ds *DataSource) ConfigString(name string) string {
	return os.Getenv(fmt.Sprintf("BAHRAM_%s", name))
}

func (ds *DataSource) ConfigByteArray(name string) []byte {
	base64Value := ds.ConfigString(name)
	value, err := base64.StdEncoding.DecodeString(base64Value)
	if err != nil {
		logging.Log(debugTag, "Error while decoding config value %s: %s", name, err)
		return nil
	}
	return value
}

func (ds *DataSource) StoreUser(u *User) error {
	userJSON, err := json.Marshal(u)
	if err != nil {
		return err
	}

	logging.Debug(debugTag, "Setting %s", userJSON)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = ds.keysAPI.Set(ctx, fmt.Sprintf("/%s/users/%s", ds.etcdDir, u.Email), string(userJSON[:]), nil)
	if err != nil {
		return err
	}
	return nil
}

func (ds *DataSource) UserByEmail(emailAddress string) (*User, error) {
	// TODO: use ds.cache

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response, err := ds.keysAPI.Get(ctx, fmt.Sprintf("/%s/users/%s", ds.etcdDir, emailAddress), nil)
	if err != nil {
		return nil, err
	}

	return userFromNodeValue(response.Node.Value)
}

func (ds *DataSource) StoreGroup(g *Group) error {
	groupJSON, err := json.Marshal(g)
	if err != nil {
		return err
	}

	logging.Debug(debugTag, "Setting %s", groupJSON)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = ds.keysAPI.Set(ctx, fmt.Sprintf("/%s/groups/%s", ds.etcdDir, g.Email), string(groupJSON[:]), nil)
	if err != nil {
		return err
	}
	return nil
}

func (ds *DataSource) GroupByEmail(emailAddress string) (*Group, error) {
	// TODO: use ds.cache

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	response, err := ds.keysAPI.Get(ctx, fmt.Sprintf("/%s/groups/%s", ds.etcdDir, emailAddress), nil)
	if err != nil {
		return nil, err
	}

	return groupFromNodeValue(response.Node.Value)
}

func (ds *DataSource) Groups() ([]*Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	response, err := ds.keysAPI.Get(ctx, fmt.Sprintf("/%s/groups", ds.etcdDir), nil)
	if err != nil {
		return nil, err
	}

	var groups []*Group

	errCount := 0
	for i := range response.Node.Nodes {
		g, e := groupFromNodeValue(response.Node.Nodes[i].Value)
		if e != nil {
			errCount += 1
			logging.Debug(debugTag, "Error while groupFromNodeValue: %s", e)
		} else {
			groups = append(groups, g)
		}
	}

	if errCount > 0 {
		return nil, fmt.Errorf("Errors happened while trying to unmarshal %d group(s)", errCount)
	}

	return groups, nil
}
