package internal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/goexl/db"
	"github.com/goexl/gox"
	"github.com/harluo/migrate/internal/internal/config"
	"github.com/harluo/migrate/internal/internal/model"
)

type Migration struct {
	dt     db.Type
	db     *sql.DB
	config *config.Migrate
}

func newMigration(dt db.Type, db *sql.DB, config *config.Migrate) *Migration {
	return &Migration{
		dt:     dt,
		db:     db,
		config: config,
	}
}

func (m *Migration) Get(ctx context.Context, migration *model.Migration) (exist bool, err error) {
	query := fmt.Sprintf(`SELECT id, version, description FROM %s WHERE id = %d`, m.config.Table, migration.Id)
	if row, qce := m.db.QueryContext(ctx, query); nil != qce {
		err = qce
	} else if nil != row && row.Next() {
		exist = true
		err = row.Scan(&migration.Id, &migration.Version, &migration.Description)
	}

	return
}

func (m *Migration) Tx(callback gox.Callback[sql.Tx]) (err error) {
	if tx, be := m.db.Begin(); nil != be {
		err = be
	} else {
		err = m.execTx(tx, callback)
	}

	return
}

func (m *Migration) execTx(tx *sql.Tx, callback gox.Callback[sql.Tx]) (err error) {
	defer func() {
		if recovered := recover(); nil != recovered {
			err = tx.Rollback()
		}
	}()

	if ce := callback(tx); nil != ce {
		err = tx.Rollback()
	} else {
		err = tx.Commit()
	}

	return
}
