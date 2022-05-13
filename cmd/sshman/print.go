package sshman

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/sonnt85/sshman"
)

//func printErrorWithHelp(c *cli.Context, err error) error {
//	cli.ShowSubcommandHelp(c)
//	fmt.Println()
//	return cli.NewExitError(err, 1)
//}

func printHosts(showPath bool, hosts []*sshman.HostConfig) {
	if DisablePrintHost {
		return
	}
	var aliases []string
	var noConnectAliases []string
	hostMap := map[string]*sshman.HostConfig{}

	for _, host := range hosts {
		hostMap[host.Alias] = host
		if host.Display() {
			aliases = append(aliases, host.Alias)
		} else {
			noConnectAliases = append(noConnectAliases, host.Alias)
		}
	}

	sort.Strings(aliases)
	for _, alias := range aliases {
		printHost(showPath, hostMap[alias])
	}

	sort.Strings(noConnectAliases)
	for _, alias := range noConnectAliases {
		printHost(showPath, hostMap[alias])
	}
}

func printHost(showPath bool, host *sshman.HostConfig) {

	if host == nil || DisablePrintHost {
		return
	}
	fmt.Printf("\t%s", color.MagentaString(host.Alias))
	if showPath && len(host.PathMap) > 0 {

		var paths []string
		for path := range host.PathMap {
			if homeDir := sshman.GetHomeDir(); strings.HasPrefix(path, homeDir) {
				path = strings.Replace(path, homeDir, "~", 1)
			}
			paths = append(paths, path)
		}
		sort.Strings(paths)
		fmt.Printf("(%s)", strings.Join(paths, " "))
	}
	if connect := host.ConnectionStr(); connect != "" {
		fmt.Printf(" -> %s", connect)
	}
	fmt.Println()
	for _, key := range sshman.SortKeys(host.OwnConfig) {
		value := host.OwnConfig[key]
		if value == "" {
			continue
		}
		color.Cyan("\t    %s = %s\n", key, value)
	}
	for _, key := range sshman.SortKeys(host.ImplicitConfig) {
		value := host.ImplicitConfig[key]
		if value == "" {
			continue
		}
		fmt.Printf("\t    %s = %s\n", key, value)
	}
	fmt.Println()
}
