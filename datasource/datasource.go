package datasource

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

func (ds *EtcdDataSource) CreateUser(active bool, values map[string]string) (User, error) {

	email, ok := values["email"]
	if !ok {
		return nil, errors.New("email is a required field in the values")
	}

	uid, ok := values["uid"]
	if !ok {
		return nil, errors.New("uid is a required field in the values")
	}

	inboxAddress, ok := values["inboxAddress"]
	if !ok {
		return nil, errors.New("inboxAddress is a required field in the values")
	}

	u := &userImpl{
		Email:        email,
		UIDStr:       uid,
		InboxAddr:    inboxAddress,
		Active:       active,
		EnFirstName:  values["enFirstName"],
		EnLastName:   values["enLastName"],
		FaFirstName:  values["faFirstName"],
		FaLastName:   values["faLastName"],
		MobileNum:    values["mobileNum"],
		EmergencyNum: values["emergencyNum"],
	}

	var err error

	birthDateStr, ok := values["birthDate"]
	if ok {
		u.BirthDate, err = strconv.ParseUint(birthDateStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error while parsing birthDate: %s", err)
		}
	}

	enrolmentDateStr, ok := values["enrolmentDate"]
	if ok {
		u.EnrolmentDate, err = strconv.ParseUint(enrolmentDateStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error while parsing enrolmentDate: %s", err)
		}
	}

	leavingDateStr, ok := values["leavingDate"]
	if ok {
		u.LeavingDate, err = strconv.ParseUint(leavingDateStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error while parsing leavingDate: %s", err)
		}
	}

	userJSON, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	logging.Debug(debugTag, "Setting %s", userJSON)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = ds.keysAPI.Set(ctx, fmt.Sprintf("/%s/users/%s", ds.etcdDir, email), string(userJSON[:]), nil)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (ds *EtcdDataSource) UserByEmail(emailAddress string) (User, error) {
	// TODO: use ds.cache

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response, err := ds.keysAPI.Get(ctx, fmt.Sprintf("/%s/users/%s", ds.etcdDir, emailAddress), nil)
	if err != nil {
		return nil, err
	}

	return userFromNodeValue(response.Node.Value)
}

func (ds *EtcdDataSource) GroupByEmail(emailAddress string) (Group, error) {
	return nil, nil
}
