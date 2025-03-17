package repository

import (
	"context"

	"github.com/necroskillz/config-service/constants"
	"gorm.io/gorm"
)

type UnitOfWorkCreator interface {
	Begin(ctx context.Context) (UnitOfWork, error)
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

type UnitOfWork interface {
	Commit() error
	Rollback() error
	Context() context.Context
}

type GormUnitOfWorkCreator struct {
	db *gorm.DB
}

type GormUnitOfWork struct {
	tx  *gorm.DB
	ctx context.Context
}

func NewGormUnitOfWork(ctx context.Context, tx *gorm.DB) UnitOfWork {
	unitOfWork := &GormUnitOfWork{tx: tx}

	ctx = context.WithValue(ctx, constants.UnitOfWorkKey, unitOfWork)

	unitOfWork.ctx = ctx

	return unitOfWork
}

func NewGormUnitOfWorkCreator(db *gorm.DB) UnitOfWorkCreator {
	return &GormUnitOfWorkCreator{db: db}
}

func (u *GormUnitOfWorkCreator) Begin(ctx context.Context) (UnitOfWork, error) {
	tx := u.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return NewGormUnitOfWork(ctx, tx), nil
}

func (u *GormUnitOfWorkCreator) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		unitOfWork := NewGormUnitOfWork(ctx, tx)

		return fn(unitOfWork.Context())
	})
}

func (u *GormUnitOfWork) Commit() error {
	return u.tx.Commit().Error
}

func (u *GormUnitOfWork) Rollback() error {
	return u.tx.Rollback().Error
}

func (u *GormUnitOfWork) Context() context.Context {
	return u.ctx
}
