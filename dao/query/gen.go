// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"gorm.io/gen"

	"gorm.io/plugin/dbresolver"
)

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:              db,
		FeedMode:        newFeedMode(db, opts...),
		Location:        newLocation(db, opts...),
		Province:        newProvince(db, opts...),
		Rendition:       newRendition(db, opts...),
		Surface:         newSurface(db, opts...),
		SurfaceFeedMode: newSurfaceFeedMode(db, opts...),
	}
}

type Query struct {
	db *gorm.DB

	FeedMode        feedMode
	Location        location
	Province        province
	Rendition       rendition
	Surface         surface
	SurfaceFeedMode surfaceFeedMode
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:              db,
		FeedMode:        q.FeedMode.clone(db),
		Location:        q.Location.clone(db),
		Province:        q.Province.clone(db),
		Rendition:       q.Rendition.clone(db),
		Surface:         q.Surface.clone(db),
		SurfaceFeedMode: q.SurfaceFeedMode.clone(db),
	}
}

func (q *Query) ReadDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Read))
}

func (q *Query) WriteDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Write))
}

func (q *Query) ReplaceDB(db *gorm.DB) *Query {
	return &Query{
		db:              db,
		FeedMode:        q.FeedMode.replaceDB(db),
		Location:        q.Location.replaceDB(db),
		Province:        q.Province.replaceDB(db),
		Rendition:       q.Rendition.replaceDB(db),
		Surface:         q.Surface.replaceDB(db),
		SurfaceFeedMode: q.SurfaceFeedMode.replaceDB(db),
	}
}

type queryCtx struct {
	FeedMode        *feedModeDo
	Location        *locationDo
	Province        *provinceDo
	Rendition       *renditionDo
	Surface         *surfaceDo
	SurfaceFeedMode *surfaceFeedModeDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		FeedMode:        q.FeedMode.WithContext(ctx),
		Location:        q.Location.WithContext(ctx),
		Province:        q.Province.WithContext(ctx),
		Rendition:       q.Rendition.WithContext(ctx),
		Surface:         q.Surface.WithContext(ctx),
		SurfaceFeedMode: q.SurfaceFeedMode.WithContext(ctx),
	}
}

func (q *Query) Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(func(tx *gorm.DB) error { return fc(q.clone(tx)) }, opts...)
}

func (q *Query) Begin(opts ...*sql.TxOptions) *QueryTx {
	tx := q.db.Begin(opts...)
	return &QueryTx{Query: q.clone(tx), Error: tx.Error}
}

type QueryTx struct {
	*Query
	Error error
}

func (q *QueryTx) Commit() error {
	return q.db.Commit().Error
}

func (q *QueryTx) Rollback() error {
	return q.db.Rollback().Error
}

func (q *QueryTx) SavePoint(name string) error {
	return q.db.SavePoint(name).Error
}

func (q *QueryTx) RollbackTo(name string) error {
	return q.db.RollbackTo(name).Error
}
