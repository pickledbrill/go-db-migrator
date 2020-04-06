package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	appLog "github.com/pickledbrill/go-db-migrator/src/logger"
)

type MigrateVersion struct {
	Logger *appLog.Logger
}

const MigrationFileExtension string = ".sql"

// NewMigrationFiles creates up and down files for migration.
func (version *MigrateVersion) NewMigrationFiles(path, name string) {
	if len(path) == 0 {
		version.Logger.LogError("Source path can't be null.")
	}
	if len(name) == 0 {
		version.Logger.LogError("Migration file name can't be null.")
	}

	now := time.Now()
	v := now.Unix()
	if isExist := checkVersionExist(path, v); isExist {
		version.Logger.LogError("Duplicate schema version files are found")
	}
	up, down := createNewVersion(v, name)
	upFileDir := fmt.Sprintf("%s/%s", path, up)
	downFileDir := fmt.Sprintf("%s/%s", path, down)
	upFile, upErr := os.Create(upFileDir)
	if upErr != nil {
		version.Logger.LogError(upErr.Error())
	}
	downFile, downErr := os.Create(downFileDir)
	if downErr != nil {
		version.Logger.LogError(downErr.Error())
	}
	upFile.Close()
	downFile.Close()
	version.Logger.LogInfo("New migration files are created.")
}

// SortVersions sorts migration files based on their versions.
func (version *MigrateVersion) SortVersions(direction string, names []string) []string {
	if len(names) == 0 {
		return nil
	}
	var versions []string
	for i, v := range names {
		index := strings.Index(v, "_")
		if index < 0 {
			continue
		}
		versions = append(versions, names[i])
	}
	switch direction {
	case "asc":
		sort.Slice(versions, func(i, j int) bool {
			return versions[i] < versions[j]
		})
		return versions
	case "desc":
		sort.Slice(versions, func(i, j int) bool {
			return versions[i] > versions[j]
		})
		return versions
	default:
		return versions
	}
}

// CheckVersionExist function look into the `migration` folder to find
// schema files with provided version.
func checkVersionExist(path string, version int64) bool {
	versionStr := strconv.FormatInt(version, 10)
	pattern := fmt.Sprintf("%s/%s_*", path, versionStr)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		panic(err)
	}
	return matches != nil
}

// createNewVersion function creates new version for `up` and `down`
// schema change files. The {version} value is integer, and is the
// epoch time value.
func createNewVersion(version int64, name string) (string, string) {
	up := fmt.Sprintf("%d_%s.up%s", version, name, MigrationFileExtension)
	down := fmt.Sprintf("%d_%s.down%s", version, name, MigrationFileExtension)
	return up, down
}
