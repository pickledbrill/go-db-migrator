package main

import (
	"flag"
	"os"

	lib "github.com/pickledbrill/go-db-migrator/src/lib"
	appLog "github.com/pickledbrill/go-db-migrator/src/logger"
)

var migrator *lib.Migrator
var version *lib.MigrateVersion
var source *lib.MigrateSource
var logger *appLog.Logger

func init() {
	migrator = new(lib.Migrator)
	version = new(lib.MigrateVersion)
	source = new(lib.MigrateSource)
	migrator.Logger = logger
	version.Logger = logger
	migrator.MigrateVersionControl = version
}

func main() {
	var newSchema string
	var dsn string
	var migrate bool
	var rollback bool
	var targetVersion string
	var init bool
	var showHistory bool

	flag.BoolVar(&init, "init", false, "Initialize migration.")
	flag.BoolVar(&migrate, "migrate", false, "Run migration or not.")
	flag.BoolVar(&rollback, "rollback", false, "Run migration rollback.")
	flag.BoolVar(&showHistory, "showHistory", false, "Show migration history.")
	flag.StringVar(&newSchema, "newSchema", "", "Create new database schema files.")
	flag.StringVar(&dsn, "dsn", os.Getenv("dsn"), "Data source name.")
	flag.StringVar(&targetVersion, "version", "", "Target version for migration or rollback.")

	flag.Parse()

	// current solution
	source.Type = "file"
	source.Location = "../migrations"
	migrator.MigrateSourceControl = source

	if dsn == "" {
		logger.LogError("Please provide DSN.")
	}
	migrator.DSN = dsn

	if init {
		migrator.InitMigration(migrator.MigrateSourceControl.Location)
		return
	}

	if len(newSchema) != 0 {
		parsedName := lib.ParseMigrationFileName(newSchema)
		version.NewMigrationFiles(migrator.MigrateSourceControl.Location, parsedName)
		return
	}

	if migrate && targetVersion == "" {
		migrator.Migrate(migrator.MigrateSourceControl.Location)
		return
	}

	if migrate && targetVersion != "" {
		migrator.MigrateTo(targetVersion)
		return
	}

	if showHistory {
		migrator.ListMigrations()
		return
	}
}
