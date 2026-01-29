package repository

import (
	"github.com/gadhittana01/go-modules-dependencies/utils"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Querier

	WithTx(tx pgx.Tx) Querier
	GetDB() utils.PGXPool
}

type RepositoryImpl struct {
	db utils.PGXPool
	*Queries
}

func NewRepository(db utils.PGXPool) Repository {
	return &RepositoryImpl{db: db, Queries: New(db)}
}

func (r *RepositoryImpl) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		db: tx,
	}
}

func (r *RepositoryImpl) GetDB() utils.PGXPool {
	return r.db
}
