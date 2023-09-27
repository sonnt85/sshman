package sshman

import (
	"fmt"
	"os"
	"strings"

	"github.com/sonnt85/sshman"
	"github.com/spf13/cobra"
)

var sshManCmd = &cobra.Command{
	Use:   "sshman",
	Short: "manage ssh_conf",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

	},
}

func GetCmd_sshMan() *cobra.Command {
	return sshManCmd
}

func Init_SshMan(parent ...*cobra.Command) {
	if len(parent) != 0 {
		parent[0].AddCommand(sshManCmd)
		// sshManCmd = parent[0]
	}
	sshmanAdd := &cobra.Command{
		Use:     "add",
		Short:   "Add a new ssh alias record",
		RunE:    addCmd,
		Aliases: []string{"a"},
	}

	sshManCmd.PersistentFlags().StringVarP(&path, "file", "f", fmt.Sprintf("%s/.ssh/config", sshman.GetHomeDir()), "Path ssh_config file")
	m := make(map[string]string)
	sshmanAdd.Flags().StringToStringP("config", "c", m, "config map[string]string")
	sshmanAdd.Flags().StringP("identityfile", "i", "", "identityfile file")
	sshmanAdd.Flags().StringP("addpath", "a", os.Getenv("MANSSH_ADD_PATH"), "addpath")
	pathShow := false
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "MANSSH_SHOW_PATH") {
			pathShow = true
			break
		}
	}
	sshmanAdd.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	sshManCmd.AddCommand(sshmanAdd)

	sshmanList := &cobra.Command{
		Use:     "list",
		Short:   "List all or query ssh alias records",
		RunE:    listCmd,
		Aliases: []string{"l"},
	}
	sshmanList.Flags().BoolP("ignorecase", "I", true, "ignore case while searching")
	sshmanList.Flags().BoolP("onname", "n", false, "Show only name alias")
	sshmanList.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	sshManCmd.AddCommand(sshmanList)

	sshmanGetOpt := &cobra.Command{
		Use:     "get",
		Short:   "Get opt of first alias  match",
		RunE:    getoptCmd,
		Aliases: []string{"g"},
	}
	sshmanGetOpt.Flags().BoolP("ignorecase", "I", true, "ignore case while searching")
	//	mansshList.Flags().BoolP("onname", "n", false, "Show only name alias")
	//	mansshList.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	sshManCmd.AddCommand(sshmanGetOpt)

	sshmanUpdate := &cobra.Command{
		Use:     "update",
		Short:   "Update the specified ssh alias",
		RunE:    updateCmd,
		Aliases: []string{"u"},
	}
	m = make(map[string]string)

	sshmanUpdate.Flags().StringToStringP("config", "c", m, "config map[string]string")
	sshmanUpdate.Flags().StringP("rename", "r", "", "rename alias")
	sshmanUpdate.Flags().StringP("addpath", "a", os.Getenv("MANSSH_ADD_PATH"), "addpath")

	sshmanUpdate.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	sshManCmd.AddCommand(sshmanUpdate)

	sshmanDelete := &cobra.Command{
		Use:     "delete",
		Short:   "Delete one or more ssh aliases",
		RunE:    deleteCmd,
		Aliases: []string{"d"},
	}

	sshmanDelete.Flags().BoolP("pathshow", "p", pathShow, "display the file path of the alias")
	sshManCmd.AddCommand(sshmanDelete)

	sshmanBackup := &cobra.Command{
		Use:     "backup",
		Short:   "Backup SSH config files",
		RunE:    backupCmd,
		Aliases: []string{"b"},
	}
	sshManCmd.AddCommand(sshmanBackup)
}

func Execute(args ...string) {
	if len(args) != 0 {
		sshManCmd.SetArgs(args)
	}
	if err := sshManCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
