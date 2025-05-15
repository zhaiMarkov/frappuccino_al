package repository

import (
	"database/sql"
	"frappuchino/internal/apperrors"
	"log/slog"
)

// checkRowsAffected проверяет, сколько строк было затронуто запросом
// и возвращает ошибку, если не было затронуто ни одной строки.
func checkRowsAffected(result sql.Result, id interface{}) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("Repository error from check row affected: failed to get rows affected", "id", id, "error", err)
		return err
	}

	if rowsAffected == 0 {
		slog.Warn("Repository info: no rows affected for ID", "id", id)
		return apperrors.ErrNotExistConflict
	}

	slog.Info("Repository info: rows affected", "id", id, "rows affected", rowsAffected)
	return nil
}
