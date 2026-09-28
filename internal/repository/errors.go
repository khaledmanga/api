package repository

import "errors"

var ErrNotFound = errors.New("resource not found")

func requireAffectedRows(result interface{ RowsAffected() (int64, error) }) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
