package sshman

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	aliasOrg, err := ListMatchFirstAlias(true, []string{`22`})
	require.Nil(t, err)
	fmt.Println(aliasOrg)
	aliasOrg, err = GetOption(aliasOrg, "port")
	require.Nil(t, err)
	fmt.Println(aliasOrg)
}
