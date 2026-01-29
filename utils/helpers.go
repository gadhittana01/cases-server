package utils

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func GetNullableString(s pgtype.Text) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func ToPgtypeTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func GetStringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
