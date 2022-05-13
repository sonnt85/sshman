package sshman

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"bytes"

	"github.com/sonnt85/sshman/sshconfig"
)

func writeConfig(p string, cfg *sshconfig.Config) error {

	oldContents := []byte{}
	if _, err := os.Stat(p); err == nil {
		oldContents, _ = ioutil.ReadFile(p)
	}

	if !bytes.Equal(oldContents, []byte(cfg.String())) {
		return ioutil.WriteFile(p, []byte(cfg.String()), 0644)
	} else {
		return nil
	}
}

func readFile(p string) (*sshconfig.Config, error) {
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return sshconfig.Decode(f)
}

func deleteHostFromConfig(config *sshconfig.Config, host *sshconfig.Host) {
	var hs []*sshconfig.Host
	for _, h := range config.Hosts {
		if h == host {
			continue
		}
		hs = append(hs, h)
	}
	config.Hosts = hs
}

func setImplicitConfig(aliasMap map[string]*HostConfig, hc *HostConfig) {
	for alias, host := range aliasMap {
		if alias == hc.Alias {
			continue
		}

		if len(hc.OwnConfig) == 0 {
			if match, err := path.Match(host.Alias, hc.Alias); err != nil || !match {
				continue
			}
			for k, v := range host.OwnConfig {
				if _, ok := hc.ImplicitConfig[k]; !ok {
					hc.ImplicitConfig[k] = v
				}
			}
			continue
		}
		if match, err := path.Match(hc.Alias, host.Alias); err != nil || !match {
			continue
		}
		for k, v := range hc.OwnConfig {
			if _, ok := host.OwnConfig[k]; ok {
				continue
			}
			if _, ok := host.ImplicitConfig[k]; !ok {
				host.ImplicitConfig[k] = v
			}
		}
	}
}

func setOwnConfig(aliasMap map[string]*HostConfig, hc *HostConfig, h *sshconfig.Host) {
	if host, ok := aliasMap[hc.Alias]; ok {
		if _, ok := host.PathMap[hc.Path]; !ok {
			host.PathMap[hc.Path] = []*sshconfig.Host{}
		}
		host.PathMap[hc.Path] = append(host.PathMap[hc.Path], h)
		for k, v := range hc.OwnConfig {
			if _, ok := host.OwnConfig[k]; !ok {
				host.OwnConfig[k] = v
			}
		}
	} else {
		aliasMap[hc.Alias] = hc
	}
}

func addHosts(aliasMap map[string]*HostConfig, fp string, hosts ...*sshconfig.Host) {
	for _, host := range hosts {
		// except implicit `*`
		if len(host.Nodes) == 0 {
			continue
		}
		for _, pattern := range host.Patterns {
			alias := pattern.String()
			hc := NewHostConfig(alias, fp, host)
			setImplicitConfig(aliasMap, hc)

			for _, node := range host.Nodes {
				if kvNode, ok := node.(*sshconfig.KV); ok {
					kvNode.Key = strings.ToLower(kvNode.Key)
					if _, ok := hc.ImplicitConfig[kvNode.Key]; !ok {
						hc.OwnConfig[kvNode.Key] = kvNode.Value
					}
				}
			}

			setImplicitConfig(aliasMap, hc)
			setOwnConfig(aliasMap, hc, host)
		}
	}
}

// ParseConfig parse configs from ssh config file, return config object and alias map
func parseConfig(p string) (map[string]*sshconfig.Config, map[string]*HostConfig, error) {
	cfg, err := readFile(p)
	if err != nil {
		return nil, nil, err
	}

	aliasMap := map[string]*HostConfig{}
	configMap := map[string]*sshconfig.Config{p: cfg}

	for _, host := range cfg.Hosts {
		var hasKV = false
		for _, node := range host.Nodes {
			switch t := node.(type) {
			case *sshconfig.Include:
				for fp, config := range t.GetFiles() {
					configMap[fp] = config
					addHosts(aliasMap, fp, config.Hosts...)
				}
			case *sshconfig.KV:
				hasKV = true
			}
		}
		if hasKV {
			addHosts(aliasMap, p, host)
		}
	}
	addHosts(aliasMap, p, &sshconfig.Host{
		Patterns: []*sshconfig.Pattern{(&sshconfig.Pattern{}).SetStr("*")},
		Nodes: []sshconfig.Node{
			sshconfig.NewKV("user", GetUsername()),
			sshconfig.NewKV("port", "22"),
		},
	})
	return configMap, aliasMap, nil
}

// ListOption options for List
type ListOption struct {
	// Keywords set Keyword filter records
	Keywords []string
	// IgnoreCase ignore case
	IgnoreCase bool
}

// List ssh alias, filter by optional keyword
func List(p string, lo ListOption) ([]*HostConfig, error) {
	configMap, aliasMap, err := parseConfig(p)
	if err != nil {
		return nil, err
	}

	var result []*HostConfig
	for _, host := range aliasMap {
		values := []string{host.Alias}
		for _, v := range host.OwnConfig {
			values = append(values, v)
		}

		if len(lo.Keywords) > 0 && !Query(values, lo.Keywords, lo.IgnoreCase) {
			continue
		}
		result = append(result, host)
	}

	// Format
	for fp, cfg := range configMap {
		if len(cfg.Hosts) > 0 {
			if err := writeConfig(fp, cfg); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

// AddOption options for Add
type AddOption struct {
	// Path add path
	Path string
	// Alias alias
	Alias string
	// Connect connection string
	Connect string
	// Config other config
	Config map[string]string
}

// Add ssh host config to ssh config file
func Add(p string, ao *AddOption) (*HostConfig, error) {
	if ao.Path == "" {
		ao.Path = p
	}

	configMap, aliasMap, err := parseConfig(p)
	if err != nil {
		return nil, err
	}
	if err := checkAlias(aliasMap, false, ao.Alias); err != nil {
		return nil, err
	}

	cfg, ok := configMap[ao.Path]
	if !ok {
		cfg, err = readFile(ao.Path)
		if err != nil {
			return nil, err
		}
	}

	// Parse connect string
	user, hostname, port := ParseConnect(ao.Connect)
	if user != "" {
		ao.Config["user"] = user
	}
	if hostname != "" {
		ao.Config["hostname"] = hostname
	}
	if port != "" {
		ao.Config["port"] = port
	}

	var nodes []sshconfig.Node
	for k, v := range ao.Config {
		nodes = append(nodes, sshconfig.NewKV(strings.ToLower(k), v))
	}

	pattern, err := sshconfig.NewPattern(ao.Alias)
	if err != nil {
		return nil, err
	}

	cfg.Hosts = append(cfg.Hosts, &sshconfig.Host{
		Patterns: []*sshconfig.Pattern{pattern},
		Nodes:    nodes,
	})
	writeConfig(ao.Path, cfg)

	_, aliasMap, err = parseConfig(p)
	if err != nil {
		return nil, err
	}
	return aliasMap[ao.Alias], nil
}

// UpdateOption options for Update
type UpdateOption struct {
	// Alias alias
	Alias string
	// NewAlias new alias
	NewAlias string
	// Connect connection string
	Connect string
	// Config other config
	Config map[string]string
}

// Valid whether the option is valid
func (uo *UpdateOption) Valid() bool {
	return uo.NewAlias != "" || uo.Connect != "" || len(uo.Config) > 0
}

// Update existing record
func Update(p string, uo *UpdateOption) (*HostConfig, error) {
	configMap, aliasMap, err := parseConfig(p)
	if err != nil {
		return nil, err
	}
	if err := checkAlias(aliasMap, true, uo.Alias); err != nil {
		return nil, err
	}

	updateHost := aliasMap[uo.Alias]
	if uo.NewAlias != "" {
		// new alias should not exist
		if err := checkAlias(aliasMap, false, uo.NewAlias); err != nil {
			return nil, err
		}
	} else {
		uo.NewAlias = uo.Alias
	}

	if uo.Connect != "" {
		// Parse connect string
		user, hostname, port := ParseConnect(uo.Connect)
		if user != "" {
			uo.Config["user"] = user
		}
		if hostname != "" {
			uo.Config["hostname"] = hostname
		}
		if port != "" {
			uo.Config["port"] = port
		}
	}

	for k, v := range uo.Config {
		if v == "" {
			delete(updateHost.OwnConfig, k)
		} else {
			updateHost.OwnConfig[k] = v
		}
	}

	for fp, hosts := range updateHost.PathMap {
		for i, host := range hosts {
			if fp == updateHost.Path {
				pattern, _ := sshconfig.NewPattern(uo.NewAlias)
				newHost := &sshconfig.Host{
					Patterns: []*sshconfig.Pattern{pattern},
				}
				for k, v := range updateHost.OwnConfig {
					newHost.Nodes = append(newHost.Nodes, sshconfig.NewKV(k, v))
				}
				if len(host.Patterns) == 1 {
					if i == 0 {
						*host = *newHost
						// for implicit "*"
						find := false
						for _, h := range configMap[fp].Hosts {
							if host == h {
								find = true
								break
							}
						}
						if !find {
							newHost.Nodes = []sshconfig.Node{}
							for k, v := range uo.Config {
								newHost.Nodes = append(newHost.Nodes, sshconfig.NewKV(k, v))
							}
							configMap[fp].Hosts = append(configMap[fp].Hosts, newHost)
						}
					} else {
						deleteHostFromConfig(configMap[fp], host)
					}
				} else {
					if i == 0 {
						configMap[fp].Hosts = append(configMap[fp].Hosts, newHost)
					}
					var patterns []*sshconfig.Pattern
					for _, pattern := range host.Patterns {
						if pattern.String() != uo.NewAlias {
							patterns = append(patterns, pattern)
						}
					}
					host.Patterns = patterns
				}
			} else {
				if len(host.Patterns) == 1 {
					deleteHostFromConfig(configMap[fp], host)
				} else {
					var patterns []*sshconfig.Pattern
					for _, pattern := range host.Patterns {
						if pattern.String() != uo.NewAlias {
							patterns = append(patterns, pattern)
						}
					}
					host.Patterns = patterns
				}
			}
		}
		if err := writeConfig(fp, configMap[fp]); err != nil {
			return nil, err
		}
	}
	_, aliasMap, err = parseConfig(p)
	if err != nil {
		return nil, err
	}
	return aliasMap[uo.NewAlias], nil
}

// Delete existing alias record
func Delete(p string, aliases ...string) ([]*HostConfig, error) {
	configMap, aliasMap, err := parseConfig(p)
	if err != nil {
		return nil, err
	}
	if err := checkAlias(aliasMap, true, aliases...); err != nil {
		return nil, err
	}

	var deleteHosts []*HostConfig
	for _, alias := range aliases {
		deleteHost := aliasMap[alias]
		deleteHosts = append(deleteHosts, deleteHost)
		for fp, hosts := range deleteHost.PathMap {
			for _, host := range hosts {
				if len(host.Patterns) == 1 {
					deleteHostFromConfig(configMap[fp], host)
				} else {
					var patterns []*sshconfig.Pattern
					for _, pattern := range host.Patterns {
						if pattern.String() != alias {
							patterns = append(patterns, pattern)
						}
					}
					host.Patterns = patterns
				}
			}
			if err := writeConfig(fp, configMap[fp]); err != nil {
				return nil, err
			}
		}
	}

	return deleteHosts, nil
}

// GetFilePaths get file paths
func GetFilePaths(p string) ([]string, error) {
	configMap, _, err := parseConfig(p)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(configMap))
	for path := range configMap {
		paths = append(paths, path)
	}
	return paths, nil
}

func checkAlias(aliasMap map[string]*HostConfig, expectExist bool, aliases ...string) error {
	for _, alias := range aliases {
		ok := aliasMap[alias] != nil
		if !ok && expectExist {
			return fmt.Errorf("alias[%s] not found", alias)
		} else if ok && !expectExist {
			return fmt.Errorf("alias[%s] already exists", alias)
		}
	}
	return nil
}
