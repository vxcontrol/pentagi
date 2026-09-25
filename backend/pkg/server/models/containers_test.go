package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rows are keyed by the type whose Valid they call.
func TestContainers_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	modelsCheckValid(t, []modelsValidCase{
		{"a starting container status", ContainerStatusStarting, ""},
		{"a running container status", ContainerStatusRunning, ""},
		{"a stopped container status", ContainerStatusStopped, ""},
		{"a deleted container status", ContainerStatusDeleted, ""},
		{"a failed container status", ContainerStatusFailed, ""},
		{"an empty container status", ContainerStatus(""), "invalid ContainerStatus: "},
		{"an unknown container status", ContainerStatus("restarting"), "invalid ContainerStatus: restarting"},

		{"a primary container type", ContainerTypePrimary, ""},
		{"a secondary container type", ContainerTypeSecondary, ""},
		{"an empty container type", ContainerType(""), "invalid ContainerType: "},
		{"an unknown container type", ContainerType("tertiary"), "invalid ContainerType: tertiary"},

		{"a complete container", modelsContainer(), ""},
		{"a container without a name", modelsWith(modelsContainer(), func(c *Container) { c.Name = "" }), "Container.Name:required"},
		{"a container without an image", modelsWith(modelsContainer(), func(c *Container) { c.Image = "" }),
			"Container.Image:required"},
		{"a container in an unknown status", modelsWith(modelsContainer(), func(c *Container) { c.Status = "invalid" }),
			"Container.Status:valid"},
		{"a container of an unknown type", modelsWith(modelsContainer(), func(c *Container) { c.Type = "invalid" }),
			"Container.Type:valid"},
	})
}

func TestContainers_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "containers", (&Container{}).TableName())
}
