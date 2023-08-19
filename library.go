package slogparquet

import (
	"runtime/debug"
	"strings"
)

const library = "samber/slog-parquet"
const goModulePath = "github.com/" + library

var semver = "v0.0.0"
var buildSHA1 = "xxxxx"

func init() {
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	for _, mod := range build.Deps {
		if mod.Replace == nil && mod.Path == goModulePath {
			semver, _, buildSHA1 = parseModuleVersion(mod.Version)
			break
		}
	}
}

func parseModuleVersion(version string) (semver, datetime, buildsha string) {
	semver, version = splitModuleVersion(version)
	datetime, version = splitModuleVersion(version)
	buildsha, _ = splitModuleVersion(version)
	semver = strings.TrimPrefix(semver, "v")
	return
}

func splitModuleVersion(s string) (head, tail string) {
	if i := strings.IndexByte(s, '-'); i < 0 {
		head = s
	} else {
		head, tail = s[:i], s[i+1:]
	}

	return
}
