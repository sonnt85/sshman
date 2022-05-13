package sshman

import (
	"fmt"
	"os"
	"strings"

	"github.com/sonnt85/sshman"
	"github.com/spf13/cobra"
)

var SshManCmd = &cobra.Command{
	Use:   "sshman",
	Short: "manage ssh_conf",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

	},
}

func Tinit() {
	mansshAdd := &cobra.Command{
		Use:     "add",
		Short:   "Add a new ssh alias record",
		RunE:    addCmd,
		Aliases: []string{"a"},
	}

	SshManCmd.PersistentFlags().StringVarP(&path, "file", "f", fmt.Sprintf("%s/.ssh/config", sshman.GetHomeDir()), "Path ssh_config file")
	m := make(map[string]string)
	mansshAdd.Flags().StringToStringP("config", "c", m, "config map[string]string")
	mansshAdd.Flags().StringP("identityfile", "i", "", "identityfile file")
	mansshAdd.Flags().StringP("addpath", "a", os.Getenv("MANSSH_ADD_PATH"), "addpath")
	pathShow := false
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "MANSSH_SHOW_PATH") {
			pathShow = true
			break
		}
	}
	mansshAdd.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	SshManCmd.AddCommand(mansshAdd)

	mansshList := &cobra.Command{
		Use:     "list",
		Short:   "List all or query ssh alias records",
		RunE:    listCmd,
		Aliases: []string{"l"},
	}
	mansshList.Flags().BoolP("ignorecase", "I", true, "ignore case while searching")
	mansshList.Flags().BoolP("onname", "n", false, "Show only name alias")
	mansshList.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	SshManCmd.AddCommand(mansshList)

	mansshGetOpt := &cobra.Command{
		Use:     "get",
		Short:   "Get opt of first alias  match",
		RunE:    getoptCmd,
		Aliases: []string{"g"},
	}
	mansshGetOpt.Flags().BoolP("ignorecase", "I", true, "ignore case while searching")
	//	mansshList.Flags().BoolP("onname", "n", false, "Show only name alias")
	//	mansshList.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	SshManCmd.AddCommand(mansshGetOpt)

	mansshUpdate := &cobra.Command{
		Use:     "update",
		Short:   "Update the specified ssh alias",
		RunE:    updateCmd,
		Aliases: []string{"u"},
	}
	m = make(map[string]string)

	mansshUpdate.Flags().StringToStringP("config", "c", m, "config map[string]string")
	mansshUpdate.Flags().StringP("rename", "r", "", "rename alias")
	mansshUpdate.Flags().StringP("addpath", "a", os.Getenv("MANSSH_ADD_PATH"), "addpath")

	mansshUpdate.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	SshManCmd.AddCommand(mansshUpdate)

	mansshDelete := &cobra.Command{
		Use:     "delete",
		Short:   "Delete one or more ssh aliases",
		RunE:    deleteCmd,
		Aliases: []string{"d"},
	}

	mansshDelete.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	SshManCmd.AddCommand(mansshDelete)

	mansshBackup := &cobra.Command{
		Use:     "backup",
		Short:   "Backup SSH config files",
		RunE:    backupCmd,
		Aliases: []string{"b"},
	}
	SshManCmd.AddCommand(mansshBackup)
}
