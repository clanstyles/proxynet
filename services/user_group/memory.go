package user_group

import (
	"context"
)

type MemoryDatastore struct {
	groups []*UserGroup
}

func NewMemoryDatastore() *MemoryDatastore {
	return &MemoryDatastore{}
}

func (ds MemoryDatastore) GetById(ctx context.Context, id string) (*UserGroup, error) {
	for _, group := range ds.groups {
		if group.ID == id {
			return group, nil
		}
	}

	return nil, ErrNoGroup
	// stmt, names := qb.Select("user_groups").Where(qb.Eq("id")).ToCql()

	// q := gocqlx.Query(ds.Session.Query(stmt), names).BindMap(qb.M{
	// 	"id": id,
	// })

	// var ug UserGroup
	// err := q.GetRelease(&ug)
	// switch {
	// case err == gocql.ErrNotFound:
	// 	return nil, ErrNoGroup
	// case err != nil:
	// 	return nil, errors.Wrap(err, "failed to release query")
	// }

	// return &ug, nil
}
