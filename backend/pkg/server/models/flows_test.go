package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlowStatusValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  FlowStatus
		wantErr bool
	}{
		{"valid created", FlowStatusCreated, false},
		{"valid running", FlowStatusRunning, false},
		{"valid waiting", FlowStatusWaiting, false},
		{"valid finished", FlowStatusFinished, false},
		{"valid failed", FlowStatusFailed, false},
		{"invalid empty", FlowStatus(""), true},
		{"invalid unknown", FlowStatus("unknown"), true},
		{"invalid paused", FlowStatus("paused"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.status.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid FlowStatus")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFlowStatusString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "created", FlowStatusCreated.String())
	assert.Equal(t, "running", FlowStatusRunning.String())
	assert.Equal(t, "finished", FlowStatusFinished.String())
}

func TestFlowTableName(t *testing.T) {
	t.Parallel()
	f := &Flow{}
	assert.Equal(t, "flows", f.TableName())
}

func TestCreateFlowValid(t *testing.T) {
	t.Parallel()

	t.Run("valid create flow", func(t *testing.T) {
		t.Parallel()
		cf := CreateFlow{Input: "scan target", Provider: "openai"}
		assert.NoError(t, cf.Valid())
	})

	t.Run("missing input", func(t *testing.T) {
		t.Parallel()
		cf := CreateFlow{Input: "", Provider: "openai"}
		assert.Error(t, cf.Valid())
	})

	t.Run("missing provider", func(t *testing.T) {
		t.Parallel()
		cf := CreateFlow{Input: "scan target", Provider: ""}
		assert.Error(t, cf.Valid())
	})
}

func TestPatchFlowValid(t *testing.T) {
	t.Parallel()

	t.Run("valid stop action", func(t *testing.T) {
		t.Parallel()
		pf := PatchFlow{Action: "stop"}
		assert.NoError(t, pf.Valid())
	})

	t.Run("valid finish action", func(t *testing.T) {
		t.Parallel()
		pf := PatchFlow{Action: "finish"}
		assert.NoError(t, pf.Valid())
	})

	t.Run("valid input action with input", func(t *testing.T) {
		t.Parallel()
		input := "new input"
		pf := PatchFlow{Action: "input", Input: &input}
		assert.NoError(t, pf.Valid())
	})

	t.Run("valid rename action with name", func(t *testing.T) {
		t.Parallel()
		name := "new name"
		pf := PatchFlow{Action: "rename", Name: &name}
		assert.NoError(t, pf.Valid())
	})

	t.Run("invalid action", func(t *testing.T) {
		t.Parallel()
		pf := PatchFlow{Action: "invalid"}
		assert.Error(t, pf.Valid())
	})

	t.Run("empty action", func(t *testing.T) {
		t.Parallel()
		pf := PatchFlow{Action: ""}
		assert.Error(t, pf.Valid())
	})

	t.Run("input action without input", func(t *testing.T) {
		t.Parallel()
		pf := PatchFlow{Action: "input"}
		assert.Error(t, pf.Valid())
	})

	t.Run("rename action without name", func(t *testing.T) {
		t.Parallel()
		pf := PatchFlow{Action: "rename"}
		assert.Error(t, pf.Valid())
	})
}

func TestTaskStatusValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  TaskStatus
		wantErr bool
	}{
		{"valid created", TaskStatusCreated, false},
		{"valid running", TaskStatusRunning, false},
		{"valid waiting", TaskStatusWaiting, false},
		{"valid finished", TaskStatusFinished, false},
		{"valid failed", TaskStatusFailed, false},
		{"invalid empty", TaskStatus(""), true},
		{"invalid unknown", TaskStatus("cancelled"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.status.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid TaskStatus")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskTableName(t *testing.T) {
	t.Parallel()
	tk := &Task{}
	assert.Equal(t, "tasks", tk.TableName())
}

func TestSubtaskStatusValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  SubtaskStatus
		wantErr bool
	}{
		{"valid created", SubtaskStatusCreated, false},
		{"valid running", SubtaskStatusRunning, false},
		{"valid waiting", SubtaskStatusWaiting, false},
		{"valid finished", SubtaskStatusFinished, false},
		{"valid failed", SubtaskStatusFailed, false},
		{"invalid empty", SubtaskStatus(""), true},
		{"invalid unknown", SubtaskStatus("pending"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.status.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid SubtaskStatus")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubtaskTableName(t *testing.T) {
	t.Parallel()
	s := &Subtask{}
	assert.Equal(t, "subtasks", s.TableName())
}

func TestContainerStatusValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  ContainerStatus
		wantErr bool
	}{
		{"valid starting", ContainerStatusStarting, false},
		{"valid running", ContainerStatusRunning, false},
		{"valid stopped", ContainerStatusStopped, false},
		{"valid deleted", ContainerStatusDeleted, false},
		{"valid failed", ContainerStatusFailed, false},
		{"invalid empty", ContainerStatus(""), true},
		{"invalid unknown", ContainerStatus("restarting"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.status.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid ContainerStatus")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainerTypeValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ct      ContainerType
		wantErr bool
	}{
		{"valid primary", ContainerTypePrimary, false},
		{"valid secondary", ContainerTypeSecondary, false},
		{"invalid empty", ContainerType(""), true},
		{"invalid tertiary", ContainerType("tertiary"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.ct.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid ContainerType")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainerTableName(t *testing.T) {
	t.Parallel()
	c := &Container{}
	assert.Equal(t, "containers", c.TableName())
}

func TestRoleValid(t *testing.T) {
	t.Parallel()

	t.Run("valid role", func(t *testing.T) {
		t.Parallel()
		r := Role{Name: "admin"}
		assert.NoError(t, r.Valid())
	})

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()
		r := Role{Name: ""}
		assert.Error(t, r.Valid())
	})
}

func TestRoleTableName(t *testing.T) {
	t.Parallel()
	r := &Role{}
	assert.Equal(t, "roles", r.TableName())
}

func TestPrivilegeValid(t *testing.T) {
	t.Parallel()

	t.Run("valid privilege", func(t *testing.T) {
		t.Parallel()
		p := Privilege{Name: "read"}
		assert.NoError(t, p.Valid())
	})

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()
		p := Privilege{Name: ""}
		assert.Error(t, p.Valid())
	})
}

func TestPrivilegeTableName(t *testing.T) {
	t.Parallel()
	p := &Privilege{}
	assert.Equal(t, "privileges", p.TableName())
}

func TestRolePrivilegesValid(t *testing.T) {
	t.Parallel()

	t.Run("valid role privileges", func(t *testing.T) {
		t.Parallel()
		rp := RolePrivileges{
			Privileges: []Privilege{{Name: "read"}, {Name: "write"}},
			Role:       Role{Name: "admin"},
		}
		assert.NoError(t, rp.Valid())
	})

	t.Run("invalid privilege", func(t *testing.T) {
		t.Parallel()
		rp := RolePrivileges{
			Privileges: []Privilege{{Name: ""}, {Name: "write"}},
			Role:       Role{Name: "admin"},
		}
		assert.Error(t, rp.Valid())
	})

	t.Run("invalid role", func(t *testing.T) {
		t.Parallel()
		rp := RolePrivileges{
			Privileges: []Privilege{{Name: "read"}},
			Role:       Role{Name: ""},
		}
		assert.Error(t, rp.Valid())
	})
}

func TestRolePrivilegesTableName(t *testing.T) {
	t.Parallel()
	rp := &RolePrivileges{}
	assert.Equal(t, "roles", rp.TableName())
}

func TestFlowTasksSubtasksTableName(t *testing.T) {
	t.Parallel()
	fts := &FlowTasksSubtasks{}
	assert.Equal(t, "flows", fts.TableName())
}

func TestFlowContainersTableName(t *testing.T) {
	t.Parallel()
	fc := &FlowContainers{}
	assert.Equal(t, "flows", fc.TableName())
}

func TestTaskSubtasksTableName(t *testing.T) {
	t.Parallel()
	ts := &TaskSubtasks{}
	assert.Equal(t, "tasks", ts.TableName())
}
