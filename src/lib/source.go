package lib

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// MigrateSource stores the type and location of the source of migration files.
type MigrateSource struct {
	Type     string
	Location string
}

// GetAllMigrationFilesFromLocal gets all migration files from
// local file system.
func (source *MigrateSource) GetAllMigrationFilesFromLocal() ([]string, error) {
	if source.Type == "file" {
		if len(source.Location) == 0 {
			return nil, errors.New("source path can't be null")
		}
		pattern := fmt.Sprintf("%s/*.sql", source.Location)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		return matches, nil
	}
	return nil, nil
}

// GetUpMigrationFilesFromLocal gets up migration files from local file system.
func (source *MigrateSource) GetUpMigrationFilesFromLocal() ([]string, error) {
	if source.Type == "file" {
		pattern := fmt.Sprintf("%s/*.up.sql", source.Location)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		return matches, nil
	}
	return nil, nil
}

// GetDownMigrationFilesFromLocal gets down migration files from local file system.
func (source *MigrateSource) GetDownMigrationFilesFromLocal() ([]string, error) {
	if source.Type == "file" {
		pattern := fmt.Sprintf("%s/*.down.sql", source.Location)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		return matches, nil
	}
	return nil, nil
}

// ValidateAllMigrationsUpDown checks up and down for all migration files in source location.
func (source *MigrateSource) ValidateAllMigrationsUpDown() ([]string, error) {
	files, err := source.GetAllMigrationFilesFromLocal()
	if err != nil {
		return nil, err
	}

	err = source.validateUpDown(files)
	return files, err
}

// ValidateMigrationUpDown checks up and down for specified migration files in source location.
func (source *MigrateSource) ValidateMigrationUpDown() ([]string, error) {
	files, err := source.GetUpMigrationFilesFromLocal()
	if err != nil {
		return nil, err
	}

	err = source.validateUpDown(files)
	return files, err
}

func (source *MigrateSource) validateUpDown(targetFiles []string) error {
	validateMap := make(map[string]bool)
	targetFileNames := []string{}
	fileNamesCollection := []string{}
	validateErr := errors.New("")
	delimiter := ""
	baseFiles, err := source.GetAllMigrationFilesFromLocal()
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		delimiter = "\\"
	} else {
		delimiter = "/"
	}

	// extract path
	for _, v := range targetFiles {
		index := strings.LastIndex(v, delimiter)
		fileName := v[index+1:]
		targetFileNames = append(targetFileNames, fileName)
	}
	for _, v := range baseFiles {
		index := strings.LastIndex(v, delimiter)
		fileName := v[index+1:]
		fileNamesCollection = append(fileNamesCollection, fileName)
	}
	// combine into one array
	joined := strings.Join(fileNamesCollection, ", ")
	joinedBytes := []byte(joined)
	// validate
	for _, v := range targetFileNames {
		versionIndex := strings.Index(v, "_")
		fileVersion := v[:versionIndex]

		if _, ok := validateMap[fileVersion]; ok {
			continue
		}
		upMatched, upErr := regexp.Match(fmt.Sprintf("%s_(.*).up.sql", fileVersion), joinedBytes)
		downMatched, downErr := regexp.Match(fmt.Sprintf("%s_(.*).down.sql", fileVersion), joinedBytes)
		if upErr != nil {
			validateErr = upErr
			break
		}
		if downErr != nil {
			validateErr = downErr
			break
		}
		if !upMatched {
			validateErr = fmt.Errorf("version %s is missing UP migration file", fileVersion)
			break
		}
		if !downMatched {
			validateErr = fmt.Errorf("version %s is missing DOWN migration file", fileVersion)
			break
		}
		validateMap[fileVersion] = true
	}

	if validateErr.Error() != "" {
		return validateErr
	}
	return nil
}
