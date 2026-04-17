package sshman

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	aliasOrg, err := ListMatchFirstAlias(true, []string{`22`})
	require.Nil(t, err)
	t.Logf("alias: %s", aliasOrg)
	port, err := GetOption(aliasOrg, "port")
	require.Nil(t, err)
	t.Logf("port: %s", port)
}
