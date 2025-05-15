package models

import (
	"strings"
)

// Преобразует имя в ID, заменяя пробелы на подчеркивания и приводя к нижнему регистру
func fromNameToID(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "_")
}
