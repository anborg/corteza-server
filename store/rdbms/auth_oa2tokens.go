package rdbms

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/cortezaproject/corteza-server/system/types"
)

func (s Store) convertAuthOa2tokenFilter(f types.AuthOa2tokenFilter) (query squirrel.SelectBuilder, err error) {
	query = s.authClientsSelectBuilder()

	if f.UserID > 0 {
		query = query.Where(squirrel.Eq{"tkn.rel_user": f.UserID})
	}

	return
}

func (s Store) DeleteExpiredAuthOA2Tokens(ctx context.Context) error {
	return s.execDeleteAuthOa2tokens(ctx, squirrel.Lt{"tkn.expires_at": "NOW()"})
}

func (s Store) DeleteAuthOA2TokenByCode(ctx context.Context, code string) error {
	return s.execDeleteAuthOa2tokens(ctx, squirrel.Gt{"tkn.code": code})
}

func (s Store) DeleteAuthOA2TokenByAccess(ctx context.Context, access string) error {
	return s.execDeleteAuthOa2tokens(ctx, squirrel.Gt{"tkn.access": access})
}

func (s Store) DeleteAuthOA2TokenByRefresh(ctx context.Context, refresh string) error {
	return s.execDeleteAuthOa2tokens(ctx, squirrel.Gt{"tkn.refresh": refresh})
}

func (s Store) DeleteAuthOA2TokenByUserID(ctx context.Context, userID uint64) error {
	return s.execDeleteAuthOa2tokens(ctx, squirrel.Gt{"tkn.rel_user": userID})
}
