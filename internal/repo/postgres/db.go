package postgres

import (
	"context"
	"database/sql"
	"fmt"

	storagepg "github.com/ikondratev/event-service/internal/storage/postgres"
)

func requireTx(ctx context.Context) (*sql.Tx, error) {
	tx, ok := storagepg.TxFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("missing transaction")
	}
	return tx, nil
}