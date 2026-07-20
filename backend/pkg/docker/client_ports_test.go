package docker

import (
	"io"
	"testing"

	"github.com/docker/go-connections/nat"
)

// TestDeterministicPortsCollide documents why container host ports are no longer
// derived from the flow id: the derivation wraps modulo
// limitContainerPortsNumber, so two flows whose ids differ by
// limitContainerPortsNumber/containerPortsNumber map to the same host ports and
// the second container's bind fails with "port is already allocated".
func TestDeterministicPortsCollide(t *testing.T) {
	period := int64(limitContainerPortsNumber / containerPortsNumber)

	a := GetPrimaryContainerPorts(2)
	b := GetPrimaryContainerPorts(2 + period)

	if !slicesEqual(a, b) {
		t.Fatalf("expected ids 2 and %d to collide, got %v vs %v", 2+period, a, b)
	}
	// 90002 ≡ 2 (mod period) → both derive to 28004/28005.
	if got := GetPrimaryContainerPorts(90002); !slicesEqual(got, GetPrimaryContainerPorts(2)) {
		t.Fatalf("expected flow 90002 to collide with flow 2, got %v", got)
	}
}

// TestReserveFreePortsCoexist is the regression: two flows created concurrently
// (ids far apart, e.g. N and N+1000) must get distinct host ports. Holding both
// reservations open at once is exactly what the RunContainer path does until it
// hands the ports to Docker, so disjointness here proves the collision is gone.
func TestReserveFreePortsCoexist(t *testing.T) {
	portsA, closersA, err := reserveFreePorts("127.0.0.1", containerPortsNumber)
	if err != nil {
		t.Fatalf("reserve A: %v", err)
	}
	defer closeAll(closersA)

	portsB, closersB, err := reserveFreePorts("127.0.0.1", containerPortsNumber)
	if err != nil {
		t.Fatalf("reserve B: %v", err)
	}
	defer closeAll(closersB)

	if len(portsA) != containerPortsNumber || len(portsB) != containerPortsNumber {
		t.Fatalf("expected %d ports each, got %v and %v", containerPortsNumber, portsA, portsB)
	}

	all := append(append([]int{}, portsA...), portsB...)
	seen := make(map[int]struct{}, len(all))
	for _, p := range all {
		if p <= 0 || p > 65535 {
			t.Fatalf("port %d out of range", p)
		}
		if _, dup := seen[p]; dup {
			t.Fatalf("two concurrent reservations collided on port %d (A=%v B=%v)", p, portsA, portsB)
		}
		seen[p] = struct{}{}
	}
}

func TestHostPortsFromBindings(t *testing.T) {
	portMap := nat.PortMap{
		"28005/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: "40002"}},
		"28004/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: "40001"}},
		"9/tcp":     nil, // no binding (host mode leaves these empty)
	}

	got := hostPortsFromBindings(portMap)

	want := []int{40001, 40002}
	if !slicesEqual(got, want) {
		t.Fatalf("expected sorted host ports %v, got %v", want, got)
	}
}

func closeAll(closers []io.Closer) {
	for _, c := range closers {
		_ = c.Close()
	}
}

func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
