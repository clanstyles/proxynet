package user_group

import (
	"context"

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

func (ds Datastore) GetById(ctx context.Context, id string) (*UserGroup, error) {
	stmt, names := qb.Select("user_groups").Where(qb.Eq("id")).ToCql()

	q := gocqlx.Query(ds.Session.Query(stmt), names).BindMap(qb.M{
		"id": id,
	})

	var ug UserGroup
	err := q.GetRelease(&ug)
	switch {
	case err == gocql.ErrNotFound:
		return nil, ErrNoGroup
	case err != nil:
		return nil, errors.Wrap(err, "failed to release query")
	}

	return &ug, nil
}
