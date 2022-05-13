package main

import (
	"github.com/sonnt85/sshman/cmd/sshman"
)

func main() {
	sshman.Tinit()
	sshman.SshManCmd.Execute()
}
