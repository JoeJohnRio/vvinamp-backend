package utils

import "database/sql"

func ToStringPtr(ns sql.NullString) *string {
	if (ns == sql.NullString{}) {
		// handle completely uninitialized case
		return nil
	}
	if ns.Valid {
		return &ns.String
	}
	return nil
}
