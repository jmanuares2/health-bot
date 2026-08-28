package bot

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"
)

func mealTypeLabel(mealType string) string {
	switch mealType {
	case "breakfast":
		return "Desayuno"
	case "lunch":
		return "Almuerzo"
	case "dinner":
		return "Cena"
	case "snack":
		return "Snack"
	default:
		return "Comida"
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func toPgNumeric(f float64) pgtype.Numeric {
	n := pgtype.Numeric{}
	_ = n.Scan(strconv.FormatFloat(f, 'f', -1, 64))
	return n
}
