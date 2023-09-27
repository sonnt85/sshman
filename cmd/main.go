package main

import (
	"github.com/sonnt85/sshman/cmd/sshman"
)

func main() {
	sshman.Init_SshMan()
	sshman.Execute()
}

// func Tnit() {
// 	slogrus.SetDefaultLoggerIsDiscard()
// 	sdaemon.SystemCheck()
// }
