package casbinmodel

import (
	"reflect"
	"strings"
)

//CasbinRules for storing the rules
type CasbinRules struct {
	PType string `db:"p_type"`
	V0    string `db:v0`
	V1    string `db:v1`
	V2    string `db:v2`
	V3    string `db:v3`
	V4    string `db:v4`
	V5    string `db:v5`
}

func (casbinRules *CasbinRules) String() string {
	const prefix = ", "

	var sb strings.Builder

	sb.Grow(
		len(casbinRules.PType) + len(casbinRules.V0) + len(casbinRules.V1) + len(casbinRules.V3) +
			len(casbinRules.V3) + len(casbinRules.V4) + len(casbinRules.V5),
	)
	// sb.Grow(len(CasbinRules))

	// s := reflect.ValueOf(casbinRules)
	s := reflect.ValueOf(casbinRules).Elem()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if len(f.Interface().(string)) > 0 {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(f.Interface().(string))
		}
	}
	return sb.String()

}
