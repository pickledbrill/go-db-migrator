package lib

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	appLog "github.com/pickledbrill/go-db-migrator/src/logger"
)

type Migrator struct {
	DSN                   string
	Logger                *appLog.Logger
	MigrateVersionControl *MigrateVersion
	MigrateSourceControl  *MigrateSource
}

// MigrationInfoTable is the default database name to store
// migraiton information.
const MigrationInfoTable string = "_MigrationInfo"

var dbQuery *DbQuery = new(DbQuery)

// InitMigration initialize the database for migration.
func (migrator *Migrator) InitMigration(path string) {
	db, err := sql.Open("mysql", migrator.DSN)
	defer db.Close()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	// create migration info table if not exist
	initQuery := dbQuery.PrepareInitQuery(MigrationInfoTable)
	stmt, err := db.Prepare(initQuery)
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	if _, err = stmt.Exec(); err != nil {
		migrator.Logger.LogError(err.Error())
	}
	// check migration history in target database
	history, err := checkMigrationHistory(db)
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	if len(history) > 0 {
		migrator.Logger.LogError("Seems like the target database already have migrations. Please reset the database")
	}
	// get all migration files, and apply migrations to target database
	versions, err := migrator.MigrateSourceControl.ValidateMigrationUpDown()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	for _, v := range versions {
		if migrateErr := applyMigration(db, v); migrateErr != nil {
			migrator.Logger.LogError(migrateErr.Error())
		}
		if migrateErr := updateMigrationHistory(db, MigrationInfoTable, v); migrateErr != nil {
			migrator.Logger.LogError(migrateErr.Error())
		}
	}
}

// Migrate applies all migrations.
func (migrator *Migrator) Migrate(path string) {
	versions, err := migrator.MigrateSourceControl.ValidateMigrationUpDown()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	if len(versions) == 0 {
		return
	}

	db, err := sql.Open("mysql", migrator.DSN)
	defer db.Close()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	selectQuery := dbQuery.PrepareLatestVersionQuery(MigrationInfoTable)
	rows, err := db.Query(selectQuery)
	defer rows.Close()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	var latestMigrationVersion string
	for rows.Next() {
		rows.Scan(&latestMigrationVersion)
	}
	err = rows.Err()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}

	latestIndex := 0
	for i, v := range versions {
		full := fmt.Sprintf("%s/%s.up.sql", path, latestMigrationVersion)
		match, err := regexp.Match(full, []byte(v))
		if err != nil {
			migrator.Logger.LogError(err.Error())
		}
		if match {
			latestIndex = i
			break
		}
	}

	migrationCandidates := versions[latestIndex+1:]

	for _, v := range migrationCandidates {
		if migrateErr := applyMigration(db, v); migrateErr != nil {
			migrator.Logger.LogError(migrateErr.Error())
		}
		if migrateErr := updateMigrationHistory(db, MigrationInfoTable, v); migrateErr != nil {
			migrator.Logger.LogError(migrateErr.Error())
		}
	}
}

// MigrateTo applies specific migration. The version should be one of
// the result of migration list.
func (migrator *Migrator) MigrateTo(version string) {
	db, err := sql.Open("mysql", migrator.DSN)
	defer db.Close()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	dbHistory, err := checkMigrationHistory(db)
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}

	exist := false
	for _, v := range dbHistory {
		if v == version {
			exist = true
			break
		}
	}
	selectQuery := dbQuery.PrepareLatestVersionQuery(MigrationInfoTable)
	rows, err := db.Query(selectQuery)
	defer rows.Close()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	var latestMigrationVersion string
	for rows.Next() {
		rows.Scan(&latestMigrationVersion)
	}
	err = rows.Err()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	var migrationSourceFiles []string
	if exist {
		if version == latestMigrationVersion {
			migrator.Logger.LogInfo(fmt.Sprintf("version %s is the latest migration history, the tool will do nothing", version))
			return
		}
		downs, err := migrator.MigrateSourceControl.GetDownMigrationFilesFromLocal()
		if err != nil {
			migrator.Logger.LogError(err.Error())
		}
		versionIndex := -1
		for i, v := range downs {
			if strings.Index(v, version) > -1 {
				versionIndex = i
				break
			}
		}
		migrationFiles := downs[versionIndex+1:]
		migrationSourceFiles = migrator.MigrateVersionControl.SortVersions("desc", migrationFiles)
	} else {
		ups, err := migrator.MigrateSourceControl.GetUpMigrationFilesFromLocal()
		if err != nil {
			migrator.Logger.LogError(err.Error())
		}
		versionIndex := -1
		dbHistoryIndex := -1
		for i, v := range ups {
			if strings.Index(v, latestMigrationVersion) > -1 {
				dbHistoryIndex = i
				continue
			}
			if strings.Index(v, version) > -1 {
				versionIndex = i
				break
			}
		}
		migrationSourceFiles = ups[dbHistoryIndex+1 : versionIndex+1]

		for _, k := range migrationSourceFiles {
			println(k)
		}
	}
	for _, v := range migrationSourceFiles {
		if migrateErr := applyMigration(db, v); migrateErr != nil {
			migrator.Logger.LogError(migrateErr.Error())
		}
		if migrateErr := updateMigrationHistory(db, MigrationInfoTable, v); migrateErr != nil {
			migrator.Logger.LogError(migrateErr.Error())
		}
	}
}

// ListMigrations retrives the migration history from the target database.
func (migrator *Migrator) ListMigrations() {
	db, err := sql.Open("mysql", migrator.DSN)
	defer db.Close()
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	history, err := checkMigrationHistory(db)
	if err != nil {
		migrator.Logger.LogError(err.Error())
	}
	for _, v := range history {
		println(v)
	}
}

// ParseMigrationFileName replaces whitespace with underscore.
func ParseMigrationFileName(name string) string {
	if len(name) == 0 {
		return ""
	}
	newName := strings.ReplaceAll(name, " ", "_")
	return newName
}

// checkMigrationHistory retrives all versions from MigrationInfo table.
func checkMigrationHistory(db *sql.DB) ([]string, error) {
	sqlQuery := dbQuery.PrepareCheckVersionQuery(MigrationInfoTable)
	rows, err := db.Query(sqlQuery)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	history := []string{}
	for rows.Next() {
		version := ""
		rows.Scan(&version)
		if len(strings.Trim(version, " ")) != 0 {
			history = append(history, version)
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return history, nil
}

func applyMigration(db *sql.DB, file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	ts, err := db.Begin()
	if err != nil {
		return err
	}
	sql := string(content)
	if _, err := ts.Exec(sql); err != nil {
		ts.Rollback()
		return err
	}
	return ts.Commit()
}

func updateMigrationHistory(db *sql.DB, tableName, file string) error {
	var query string
	if strings.Index(file, "up") > -1 {
		query = dbQuery.PrepareInsertVersionQuery(tableName)
	} else {
		query = dbQuery.PrepareDeleteVersionQuery(tableName)
	}
	// insertQuery := dbQuery.PrepareInsertVersionQuery(tableName)
	ts, err := db.Begin()
	if err != nil {
		return err
	}
	del := ""
	if runtime.GOOS == "windows" {
		del = "\\"
	} else {
		del = "/"
	}
	i := strings.LastIndex(file, del)
	version := file[i+1:]
	dotIndex := strings.Index(version, ".")
	versionName := version[:dotIndex]
	if _, err := ts.Exec(query, versionName); err != nil {
		ts.Rollback()
		return err
	}
	return ts.Commit()
}
