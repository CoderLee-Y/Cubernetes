package testing

import (
	"Cubernetes/pkg/cubenetwork/register"
	"testing"
)

func TestRegister(t *testing.T) {
	args := []string{"xx", "192.168.1.5", "192.168.1.9"}
	register.RegistryMaster(args)
}
