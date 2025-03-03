package sshman

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sonnt85/gosutils/sutils"
	"github.com/sonnt85/gosystem"
	"github.com/sonnt85/sshman"
	"github.com/spf13/cobra"
)

var (
	path             = fmt.Sprintf("%s/.ssh/config", gosystem.GetHomeDir())
	DisablePrintHost bool
)

type SshConfig struct {
	path string
}

func getArgs(index int, args []string) string {
	if len(args) > index {
		return args[index]
	} else {
		return ""
	}
}

func NewSshConfig(path string) *SshConfig {
	return &SshConfig{path: path}
}

func (sc *SshConfig) ListSSH(ign, pathShowFlag, onname bool, args []string) error {
	hosts, err := sshman.List(sc.path, sshman.ListOption{
		Keywords:   args,
		IgnoreCase: ign,
	})
	if err != nil {
		fmt.Printf(sshman.ErrorFlag)
		return err
	}
	fmt.Printf("%s total records: %d\n\n", sshman.SuccessFlag, len(hosts))
	printHosts(pathShowFlag, hosts)
	return nil
}

func ListSSH(ign, pathShowFlag, onname bool, args []string, paths ...string) error {
	cfgpath := path
	if len(paths) != 0 {
		cfgpath = paths[0]
	}
	return NewSshConfig(cfgpath).ListSSH(ign, pathShowFlag, onname, args)
}

func listCmd(c *cobra.Command, args []string) error {
	onname, _ := c.Flags().GetBool("onname")
	if onname {
		if aliaslist, err := ListMatchAlias(true, args); err == nil {
			for _, v := range aliaslist {
				if v != "*" {
					fmt.Println(v)
				}
			}
			return nil
		}
		return errors.New("no alias found")
	} else {
		ign, _ := c.Flags().GetBool("ignorecase")
		pathShowFlag, _ := c.Flags().GetBool("pathshow")
		return ListSSH(ign, pathShowFlag, onname, args)
	}
}

func getoptCmd(c *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("missing args")
	}
	ign, _ := c.Flags().GetBool("ignorecase")
	//	fmt.Println(args)
	if opt, err := GetOption(args[0], args[1], ign); err == nil {
		fmt.Println(opt)
		return nil
	} else {
		return err
	}

	//	return ListSSH(ign, pathShowFlag, onname, args)
}

func GetOption(alias, optionname string, ignorecases ...bool) (ret string, err error) {
	igncase := false
	if len(ignorecases) != 0 {
		igncase = ignorecases[0]
	}
	hosts, err := sshman.List(path, sshman.ListOption{
		Keywords:   []string{alias},
		IgnoreCase: igncase,
	})
	if err != nil {
		return "", err
	}
	host := hosts[0]

	ret, ok := host.OwnConfig[optionname]

	if ok {
		return ret, nil
	} else {
		ret, ok := host.ImplicitConfig[optionname]
		if ok {
			return ret, nil
		}
	}
	return "", errors.New("Missing key: " + optionname)
}

func ListMatchAlias(IgnoreCase bool, args []string) (alias []string, err error) {
	alias = []string{}
	hosts, err := sshman.List(path, sshman.ListOption{
		Keywords:   args,
		IgnoreCase: IgnoreCase,
	})
	if err != nil {
		return alias, err
	}
	for _, host := range hosts {
		alias = append(alias, host.Alias)
	}
	return alias, nil
}

func ListMatchFirstAlias(IgnoreCase bool, args []string) (alias string, err error) {
	aliass, err := ListMatchAlias(IgnoreCase, args)
	if err != nil {
		return alias, err
	} else if len(aliass) == 0 {
		return "", errors.New("can not find match alias")
	}
	return aliass[0], nil
}

func AddAlias(addpath, identityfile string, kvConfig map[string]string, pathShowFlag bool, args []string, disablePrints ...bool) error {
	// Check arguments count
	enablePrint := len(disablePrints) != 0 && disablePrints[0]
	enablePrint = !enablePrint
	if err := sshman.ArgumentsCheck(len(args), 1, 2); err != nil {
		return err
	}
	if addpath == "" {
		addpath = path
	}
	ao := &sshman.AddOption{
		Alias:   getArgs(0, args),
		Connect: getArgs(1, args),
		Path:    addpath,
	}
	if ao.Path != "" {
		var err error
		if ao.Path, err = filepath.Abs(ao.Path); err != nil {
			fmt.Printf(sshman.ErrorFlag)
			return err
		}
	}
	if len(kvConfig) != 0 {
		ao.Config = kvConfig
	}
	if ao.Config == nil {
		ao.Config = make(map[string]string)
	}

	if len(identityfile) != 0 {
		ao.Config["identityfile"] = identityfile
	}

	if len(ao.Config) == 0 && ao.Connect == "" {
		return errors.New("param error")
	}

	host, err := sshman.Add(path, ao)
	if err != nil {
		if enablePrint {
			fmt.Printf(sshman.ErrorFlag)
		}
		return err
	}

	if !DisablePrintHost {
		if enablePrint {
			fmt.Printf("%s added successfully\n", sshman.SuccessFlag)
			if host != nil {
				fmt.Println()
				printHost(pathShowFlag, host)
			}
		}
	}
	return nil
}

func addCmd(c *cobra.Command, args []string) error {
	addpath, _ := c.Flags().GetString("addpath")
	kvConfig, _ := c.Flags().GetStringToString("config")
	identityfile, _ := c.Flags().GetString("identityfile")
	pathShowFlag, _ := c.Flags().GetBool("pathshow")
	return AddAlias(addpath, identityfile, kvConfig, pathShowFlag, args)
}

// args[0] -> origin alias
// args[1] -> Host include ip
func UpsertSSH(remname, identityfile string, kvConfig map[string]string, pathShowFlag bool, args []string, disablePrints ...bool) error {
	aliasOrg, _ := ListMatchFirstAlias(false, []string{args[0]})
	// if err != nil {
	// 	return err
	// }
	if aliasOrg == "" {
		return AddAlias("", identityfile, kvConfig, pathShowFlag, args, disablePrints...)
	} else {
		args[0] = aliasOrg
		return UpdateSSH(remname, identityfile, kvConfig, pathShowFlag, args, disablePrints...)
	}
}

func UpdateSSH(remname, identityfile string, kvConfig map[string]string, pathShowFlag bool, args []string, disablePrints ...bool) error {
	// Check arguments count
	enablePrint := len(disablePrints) != 0 && disablePrints[0]
	enablePrint = !enablePrint
	if err := sshman.ArgumentsCheck(len(args), 1, 2); err != nil {
		return err
	}
	uo := &sshman.UpdateOption{
		Alias:    getArgs(0, args),
		Connect:  getArgs(1, args),
		NewAlias: remname,
	}
	if len(kvConfig) != 0 {
		uo.Config = kvConfig
	}
	if uo.Config == nil {
		uo.Config = make(map[string]string)
	}

	if identityfile != "" {
		uo.Config["identityfile"] = identityfile
	}
	if !uo.Valid() {
		return errors.New("the update option is invalid")
	}

	host, err := sshman.Update(path, uo)

	if err != nil {
		if enablePrint {
			fmt.Printf(sshman.ErrorFlag)
		}
		return err
	}

	if !DisablePrintHost {
		if enablePrint {
			fmt.Printf("%s updated successfully\n\n", sshman.SuccessFlag)
			printHost(pathShowFlag, host)
		}
	}
	return nil
}

func updateCmd(c *cobra.Command, args []string) error {
	remname, _ := c.Flags().GetString("rename")
	kvConfig, _ := c.Flags().GetStringToString("config")
	identityfile, _ := c.Flags().GetString("identityfile")
	pathShowFlag, _ := c.Flags().GetBool("pathshow")
	return UpdateSSH(remname, identityfile, kvConfig, pathShowFlag, args)
}

func DeleteAlias(pathShowFlag bool, args []string, disablePrints ...bool) error {
	enablePrint := len(disablePrints) != 0 && disablePrints[0]
	enablePrint = !enablePrint
	if err := sshman.ArgumentsCheck(len(args), 1, -1); err != nil {
		return err
	}
	hosts, err := sshman.Delete(path, args...)
	if err != nil {
		if enablePrint {
			fmt.Printf(sshman.ErrorFlag)
		}
		return err
	}
	if !DisablePrintHost {
		if enablePrint {
			fmt.Printf("%s deleted successfully\n\n", sshman.SuccessFlag)
			printHosts(pathShowFlag, hosts)
		}
	}
	return nil
}

func deleteCmd(c *cobra.Command, args []string) error {
	pathShowFlag, _ := c.Flags().GetBool("pathshow")
	return DeleteAlias(pathShowFlag, args)
}

func BackupSSH(args []string, disablePrints ...bool) error {
	//	fmt.Println("running backup ...")
	//	if err := sshman.ArgumentsCheck(len(args), 1, 1); err != nil {
	//		return err
	//	}
	enablePrint := len(disablePrints) != 0 && disablePrints[0]
	enablePrint = !enablePrint
	backupPath := getArgs(0, args)
	if len(backupPath) == 0 {
		backupPath = "."
	} else {
		os.MkdirAll(backupPath, os.ModePerm)
	}

	paths, err := sshman.GetFilePaths(path)
	if err != nil {
		return err
	}
	pathDir := filepath.Dir(path)
	for _, p := range paths {
		bp := backupPath
		if p != path && strings.HasPrefix(p, pathDir) {
			bp = filepath.Join(bp, strings.Replace(p, pathDir, "", 1))
			os.MkdirAll(filepath.Dir(bp), os.ModePerm)
		}
		if !sutils.IoCopy(p, filepath.Join(bp, filepath.Base(p))) {
			//		if err := exec.Command("cp", p, bp).Run(); err != nil {
			return err
		}
	}
	if enablePrint {
		fmt.Printf("%s backup ssh config to [%s] successfully\n", sshman.SuccessFlag, backupPath)
	}
	return nil
}

func backupCmd(c *cobra.Command, args []string) error {
	return BackupSSH(args)
}
