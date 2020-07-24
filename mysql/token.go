package mysql

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

type tokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) neo.TokenRepository {
	return &tokenRepository{
		db,
	}
}

func (r *tokenRepository) Token(ctx context.Context, id uint64) (*neo.Token, error) {
	var token = new(neo.Token)
	err := boiler.Tokens(
		boiler.TokenWhere.ID.EQ(id),
	).Bind(ctx, r.db, token)

	return token, err
}

func (r *tokenRepository) CreateToken(ctx context.Context, token *neo.Token) (*neo.Token, error) {

	var bToken = new(boiler.Token)
	err := copier.Copy(bToken, token)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	err = bToken.Insert(ctx, r.db, boil.Infer(), false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert token")
	}
	token = new(neo.Token)
	err = copier.Copy(token, bToken)

	return token, errors.Wrap(err, "failed to copy token")
}

func (r *tokenRepository) UpdateToken(ctx context.Context, id uint64, token *neo.Token) (*neo.Token, error) {

	var bToken = new(boiler.Token)
	err := copier.Copy(bToken, token)
	if err != nil {
		return nil, errors.Wrap(err, "unable to copy token")
	}

	bToken.ID = id

	_, err = bToken.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "unable to update token")
	}
	token = new(neo.Token)
	err = copier.Copy(token, bToken)

	return token, errors.Wrap(err, "failed to copy token")
}

func (r *tokenRepository) DeleteToken(ctx context.Context, id uint64) error {

	_, err := boiler.Tokens(
		boiler.TokenWhere.ID.EQ(id),
	).DeleteAll(ctx, r.db)

	return err

}
