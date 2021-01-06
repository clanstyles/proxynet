package user

import (
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

type Datastore struct {
	*gocql.Session
}

func NewDatastore(sess *gocql.Session) *Datastore {
	return &Datastore{sess}
}

func (ds Datastore) Create(username, password string) (*User, error) {
	stmt, names := qb.Insert("users").
		Columns("billing_id", "username", "passwords", "status", "ips").
		Timestamp(time.Now()).
		ToCql()

	var u User
	if err := gocqlx.Query(ds.Session.Query(stmt), names).BindStruct(&u).ExecRelease(); err != nil {
		return nil, errors.Wrap(err, "failed to query user group by id")
	}

	return &u, nil
}

func (ds Datastore) GetByUsername(username string) (*User, error) {
	stmt, names := qb.Select("users").Where(qb.Eq("username")).ToCql()

	var u User
	q := gocqlx.Query(ds.Session.Query(stmt), names).BindMap(qb.M{
		"username": username,
	})

	err := q.GetRelease(&u)
	switch {
	case err == gocql.ErrNotFound:
		return nil, ErrNoUser
	case err != nil:
		return nil, errors.Wrap(err, "failed to release query")
	}

	log.Printf("%+v", u)
	return &u, nil
}

func (ds Datastore) ValidCredentials(username, password string) (*User, error) {
	if username == "" {
		return nil, ErrInvalidCredentials
	}

	stmt, names := qb.Select("users").Where(qb.Eq("username")).ToCql()

	var u User
	q := gocqlx.Query(ds.Session.Query(stmt), names).BindMap(qb.M{
		"username": username,
	})

	err := q.GetRelease(&u)
	switch {
	case err == gocql.ErrNotFound:
		return nil, ErrInvalidCredentials

	case err != nil:
		return nil, errors.Wrap(err, "failed to release query")
	}

	if u.Password == password {
		return &u, nil
	}

	return nil, ErrInvalidCredentials
}

func (ds Datastore) GetByIP(ip string) (*User, error) {
	stmt, names := qb.Select("users_by_ip").Columns("username").Where(qb.Eq("ip")).ToCql()

	var username string
	q := gocqlx.Query(ds.Session.Query(stmt), names).BindMap(qb.M{
		"ip": ip,
	})

	err := q.GetRelease(&username)
	switch {
	case err == gocql.ErrNotFound:
		return nil, ErrNoIP

	case err != nil:
		return nil, errors.Wrap(err, "failed to release query")
	}

	return ds.GetByUsername(username)
}

func (ds Datastore) Disable(username string) error {
	return nil
}
