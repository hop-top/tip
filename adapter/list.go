package adapter

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// AdapterInfo holds discovered adapter metadata.
type AdapterInfo struct {
	Name string
	Path string
	Desc *DescribeResponse // nil if describe fails
}

// List scans $PATH and plugin dir for tip-* binaries.
// Returns deduplicated list with describe info.
func List() []AdapterInfo {
	return ListWithHome("")
}

// ListWithHome is like List but accepts an explicit home dir.
func ListWithHome(home string) []AdapterInfo {
	seen := make(map[string]bool)
	var infos []AdapterInfo

	// 1. Scan $PATH
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	for _, dir := range pathDirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasPrefix(e.Name(), "tip-") {
				continue
			}
			fi, err := e.Info()
			if err != nil || fi.Mode()&0o111 == 0 {
				continue
			}
			eName := e.Name()
			if runtime.GOOS == "windows" {
				eName = strings.TrimSuffix(eName, ".exe")
			}
			name := strings.TrimPrefix(eName, "tip-")
			if seen[name] {
				continue
			}
			seen[name] = true
			p := filepath.Join(dir, e.Name())
			infos = append(infos, AdapterInfo{Name: name, Path: p})
		}
	}

	// 2. Scan plugin dir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return infos
		}
	}
	plugDir := filepath.Join(home, PluginDir)
	entries, err := os.ReadDir(plugDir)
	if err != nil {
		return infos
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "tip-") {
			continue
		}
		name := strings.TrimPrefix(e.Name(), "tip-")
		if seen[name] {
			continue
		}
		seen[name] = true
		p := filepath.Join(plugDir, e.Name())
		infos = append(infos, AdapterInfo{Name: name, Path: p})
	}

	return infos
}

// ListWithDescribe calls List then populates describe info.
func ListWithDescribe(cache *DescribeCache) []AdapterInfo {
	return ListWithDescribeHome("", cache)
}

// ListWithDescribeHome calls ListWithHome then populates describe info.
func ListWithDescribeHome(home string, cache *DescribeCache) []AdapterInfo {
	infos := ListWithHome(home)
	for i := range infos {
		desc, _ := cache.CachedDescribe(infos[i].Path)
		infos[i].Desc = desc
	}
	return infos
}
