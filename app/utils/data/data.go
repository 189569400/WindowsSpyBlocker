package data

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/crazy-max/WindowsSpyBlocker/app/bindata"
	"github.com/crazy-max/WindowsSpyBlocker/app/utils/config"
	"github.com/crazy-max/WindowsSpyBlocker/app/utils/netu"
	"github.com/crazy-max/WindowsSpyBlocker/app/utils/pathu"
	"github.com/pkg/errors"
)

// Systems, rules, types and exts constants
const (
	OS_WIN7  = "win7"
	OS_WIN81 = "win81"
	OS_WIN10 = "win10"

	RULES_EXTRA  = "extra"
	RULES_SPY    = "spy"
	RULES_UPDATE = "update"

	TYPE_FIREWALL = "firewall"
	TYPE_HOSTS    = "hosts"

	EXT_DNSCRYPT   = "dnscrypt"
	EXT_OPENWRT    = "openwrt"
	EXT_P2P        = "p2p"
	EXT_PROXIFIER  = "proxifier"
	EXT_SIMPLEWALL = "simplewall"
)

func getAsset(assetPath string) ([]string, error) {
	if config.App.UseEmbeddedData {
		return getAssetEmbbeded(assetPath)
	} else {
		return getAssetExternal(assetPath)
	}
}

func getAssetEmbbeded(assetPath string) ([]string, error) {
	result, err := bindata.Asset(assetPath)
	if err != nil {
		return []string{}, err
	}
	return strings.Split(string(result), "\n"), nil
}

func getAssetExternal(assetPath string) ([]string, error) {
	extPath := path.Join(pathu.Current, assetPath)
	if _, err := os.Stat(extPath); err != nil {
		return []string{}, errors.New(fmt.Sprintf("Cannot stat file: %s", strings.TrimLeft(extPath, pathu.Current)))
	}

	extFile, err := os.Open(extPath)
	if err != nil {
		return []string{}, errors.New(fmt.Sprintf("Cannot open file: %s", strings.TrimLeft(extPath, pathu.Current)))
	}
	defer extFile.Close()

	extFileBuf, err := ioutil.ReadFile(extPath)
	if err != nil {
		return []string{}, errors.New(fmt.Sprintf("Cannot read file: %s", strings.TrimLeft(extPath, pathu.Current)))
	}

	return strings.Split(string(extFileBuf), "\n"), nil
}

func getIp(ip string) string {
	if strings.Contains(ip, "-") {
		ipRange := strings.SplitN(ip, "-", 2)
		if len(ipRange) != 2 {
			return ip
		}
		if !netu.IsValidIPv4(ipRange[0]) {
			return ip
		}
		return ipRange[0]
	}
	return ip
}

// GetFirewallIps returns ips filtered by system
func GetFirewallIps(system string) (ips, error) {
	var result ips

	extra, err := GetFirewallIpsByRule(system, RULES_EXTRA)
	if err != nil {
		return result, err
	}
	result = append(result, extra...)

	spy, err := GetFirewallIpsByRule(system, RULES_SPY)
	if err != nil {
		return result, err
	}
	result = append(result, spy...)

	update, err := GetFirewallIpsByRule(system, RULES_UPDATE)
	if err != nil {
		return result, err
	}
	result = append(result, update...)

	sort.Sort(result)
	return result, nil
}

// GetFirewallIpsByRule returns ips filtered by system and rule
func GetFirewallIpsByRule(system string, rule string) (ips, error) {
	var result ips

	rulesPath := path.Join("data", TYPE_FIREWALL, system, rule+".txt")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}

	if len(lines) == 0 {
		return result, fmt.Errorf("No IPs found in %s", rulesPath)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if netu.IsValidIpv4Range(line) || netu.IsValidIPv4(line) {
			result = append(result, ip{IP: line})
		}
	}

	sort.Sort(result)
	return result, nil
}

// GetHosts returns hosts filtered by system
func GetHosts(system string) (hosts, error) {
	var result hosts

	extra, err := GetHostsByRule(system, RULES_EXTRA)
	if err != nil {
		return result, err
	}
	result = append(result, extra...)

	spy, err := GetHostsByRule(system, RULES_SPY)
	if err != nil {
		return result, err
	}
	result = append(result, spy...)

	update, err := GetHostsByRule(system, RULES_UPDATE)
	if err != nil {
		return result, err
	}
	result = append(result, update...)

	sort.Sort(result)
	return result, nil
}

// GetHostsByRule returns hosts filtered by system and rule
func GetHostsByRule(system string, rule string) (hosts, error) {
	var result hosts

	rulesPath := path.Join("data", TYPE_HOSTS, system, rule+".txt")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}

	if len(lines) == 0 {
		return result, fmt.Errorf("No domains found in %s", rulesPath)
	}

	for _, line := range lines {
		line = strings.TrimRight(strings.TrimLeft(strings.TrimSpace(line), "0.0.0.0 "), ":443")
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		result = append(result, host{Domain: line})
	}

	sort.Sort(result)
	return result, nil
}

// GetExtIPs returns IPs for an external data filtered by system and rule
func GetExtIPs(ext string, system string, rule string) (ips, error) {
	var err error
	var result ips

	if ext == EXT_OPENWRT {
		result, err = getOpenwrtIPs(system, rule)
		if err != nil {
			return nil, err
		}
	} else if ext == EXT_P2P {
		result, err = getP2pIPs(system, rule)
		if err != nil {
			return nil, err
		}
	} else if ext == EXT_PROXIFIER {
		result, err = getProxifierIPs(system, rule)
		if err != nil {
			return nil, err
		}
	} else if ext == EXT_SIMPLEWALL {
		result, err = getSimplewallIPs(system, rule)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// GetExtHosts returns hosts for an external data filtered by system and rule
func GetExtHosts(ext string, system string, rule string) (hosts, error) {
	var err error
	var result hosts

	if ext == EXT_DNSCRYPT {
		result, err = getDnscryptHosts(system, rule)
		if err != nil {
			return nil, err
		}
	} else if ext == EXT_OPENWRT {
		result, err = getOpenwrtHosts(system, rule)
		if err != nil {
			return nil, err
		}
	} else if ext == EXT_PROXIFIER {
		result, err = getProxifierHosts(system, rule)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func getDnscryptHosts(system string, rule string) (hosts, error) {
	var result hosts

	rulesPath := path.Join("data", EXT_DNSCRYPT, system, rule+".txt")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}

	if len(lines) == 0 {
		return result, fmt.Errorf("No domains found in %s", rulesPath)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, host{Domain: line})
	}

	sort.Sort(result)
	return result, nil
}

func getOpenwrtIPs(system string, rule string) (ips, error) {
	var result ips

	rulesPath := path.Join("data", EXT_OPENWRT, system, rule, "firewall.user")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}

	if len(lines) == 0 {
		return result, fmt.Errorf("No IPs found in %s", rulesPath)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "iptables -I FORWARD -j DROP -d ") {
			continue
		}

		line = strings.TrimLeft(line, "iptables -I FORWARD -j DROP -d ")
		if strings.Contains(line, "/") {
			_, _, err := net.ParseCIDR(line)
			if err != nil {
				continue
			}
		}

		result = append(result, ip{IP: line})
	}

	sort.Sort(result)
	return result, nil
}

func getOpenwrtHosts(system string, rule string) (hosts, error) {
	var result hosts

	rulesPath := path.Join("data", EXT_OPENWRT, system, rule, "dnsmasq.conf")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}

	if len(lines) == 0 {
		return result, fmt.Errorf("No domains found in %s", rulesPath)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "server=/") {
			continue
		}
		lineAr := strings.Split(line, "/")
		result = append(result, host{Domain: lineAr[1]})
	}

	sort.Sort(result)
	return result, nil
}

func getProxifierIPs(system string, rule string) (ips, error) {
	var result ips

	rulesPath := path.Join("data", EXT_PROXIFIER, system, rule, "ips.txt")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}

	if len(lines) == 0 {
		return result, fmt.Errorf("No IPs found in %s", rulesPath)
	}

	for _, line := range lines {
		line = strings.TrimRight(strings.TrimSpace(line), ";")
		result = append(result, ip{IP: line})
	}

	sort.Sort(result)
	return result, nil
}

func getProxifierHosts(system string, rule string) (hosts, error) {
	var result hosts

	rulesPath := path.Join("data", EXT_PROXIFIER, system, rule, "domains.txt")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}

	if len(lines) == 0 {
		return result, fmt.Errorf("No domains found in %s", rulesPath)
	}

	for _, line := range lines {
		line = strings.TrimRight(strings.TrimSpace(line), ";")
		result = append(result, host{Domain: line})
	}

	sort.Sort(result)
	return result, nil
}

func getSimplewallIPs(system string, rule string) (ips, error) {
	var result ips

	rulesPath := path.Join("data", EXT_SIMPLEWALL, system, rule, "blocklist.xml")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}
	rules, _ := bindata.Asset(rulesPath)

	if len(lines) == 0 {
		return result, fmt.Errorf("No IPs found in %s", rulesPath)
	}

	var root SimplewallRoot
	xml.Unmarshal(rules, &root)

	for _, item := range root.ItemList {
		result = append(result, ip{IP: item.rule})
	}

	sort.Sort(result)
	return result, nil
}

func getP2pIPs(system string, rule string) (ips, error) {
	var result ips

	rulesPath := path.Join("data", EXT_P2P, system, rule+".txt")
	lines, err := getAsset(rulesPath)
	if err != nil {
		return result, err
	}

	if len(lines) == 0 {
		return result, fmt.Errorf("No IPs found in %s", rulesPath)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "WindowsSpyBlocker:") {
			continue
		}
		ipRange := strings.TrimLeft(line, "WindowsSpyBlocker:")
		lineAr := strings.Split(ipRange, "-")
		if lineAr[0] == lineAr[1] {
			result = append(result, ip{IP: lineAr[0]})
		} else {
			result = append(result, ip{IP: ipRange})
		}
	}

	sort.Sort(result)
	return result, nil
}

// GetIPsSlice returns IPs as slice
func GetIPsSlice(resultIps ips) []string {
	var result []string

	for _, resultIp := range resultIps {
		result = append(result, resultIp.IP)
	}

	return result
}

// GetHostsSlice returns hosts as slice
func GetHostsSlice(resultHosts hosts) []string {
	var result []string

	for _, resultHost := range resultHosts {
		result = append(result, resultHost.Domain)
	}

	return result
}
