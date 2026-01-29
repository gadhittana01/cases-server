package utils

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func TimeToPgtypeTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func PgtypeTimeToTime(pt pgtype.Timestamptz) time.Time {
	if !pt.Valid {
		return time.Time{}
	}
	return pt.Time
}

func UUIDToPgtypeUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: *u,
		Valid: true,
	}
}

func PgtypeUUIDToUUID(pt pgtype.UUID) *uuid.UUID {
	if !pt.Valid {
		return nil
	}
	u := uuid.UUID(pt.Bytes)
	return &u
}

func BoolToPgtypeBool(b bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  b,
		Valid: true,
	}
}

func PgtypeBoolToBool(pt pgtype.Bool) bool {
	if !pt.Valid {
		return false
	}
	return pt.Bool
}

func DecimalToPgtypeNumeric(d decimal.Decimal) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   d.Coefficient(),
		Exp:   d.Exponent(),
		NaN:   false,
		Valid: true,
	}
}

func PgtypeNumericToDecimal(pt pgtype.Numeric) *decimal.Decimal {
	if !pt.Valid {
		return nil
	}
	if pt.NaN {
		return nil
	}
	d := decimal.NewFromBigInt(pt.Int, pt.Exp)
	return &d
}
