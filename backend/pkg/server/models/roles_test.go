package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rows are keyed by the type whose Valid they call.
func TestRoles_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	admin := Role{Name: "admin"}

	modelsCheckValid(t, []modelsValidCase{
		{"a named role", admin, ""},
		{"an unnamed role", Role{}, "Role.Name:required"},

		{"a named privilege", Privilege{Name: "read"}, ""},
		{"an unnamed privilege", Privilege{}, "Privilege.Name:required"},

		{"a role with its privileges", RolePrivileges{Privileges: []Privilege{{Name: "read"}, {Name: "write"}}, Role: admin}, ""},
		{"a role with an unnamed privilege", RolePrivileges{Privileges: []Privilege{{}, {Name: "write"}}, Role: admin},
			"Privilege.Name:required"},
		{"an unnamed role with its privileges", RolePrivileges{Privileges: []Privilege{{Name: "read"}}}, "Role.Name:required"},
	})
}

func TestRoles_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model modelsTabled
		want  string
	}{
		{&Role{}, "roles"},
		{&Privilege{}, "privileges"},
		{&RolePrivileges{}, "roles"},
	} {
		assert.Equal(t, tc.want, tc.model.TableName(), "%T", tc.model)
	}
}
