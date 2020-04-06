# go-db-migrator
The tool only supports `mysql` currently, and can only be used as a command line tool. In order to check how to use the CLI, please see the `Run CLI` section.
## Run CLI
- ### Initialize Database for Migration
```
      $ .\migrate.exe -init=true -dsn="root:my-password@tcp(127.0.0.1:46985)/test"
```
This command will create `_MigrationInfo` table, and apply all the schema change from the `migrations` folder. The table `_MigrationInfo` is used to store applied migration files. The tool will use the information from this table to track schema change.
- ### Create New Migration Files
```
      $ .\migrate.exe -newSchema="test" -dsn="root:my-password@tcp(127.0.0.1:46985)/test"
```
This command will create empty database schema change files in the `migrations` folder. The `migrations` folder is under `src` directory. Click [here](#database-schema-files) to check the database schema files generated by this tool.
- ### Migrate All
```
      $ .\migrate.exe -migrate=true -dsn="root:my-password@tcp(127.0.0.1:46985)/test"
```
This command will compare the latest migration history inside the `_MigrationInfo` table, and will apply all `up` files that are newer than the history.
- ### Migrate to Specific Version
```
      $ .\migrate.exe -migrate=true -targetVersion="1585990417_test_one" -dsn="root:my-password@tcp(127.0.0.1:46985)/test"
```
This command can be used to migrate the database schema to specific version. If the version doesn't exist in the migration history table, then the tool will apply the `up` files. Otherwise it will apply the `down` files which are used for rollback.
- ### Show Migration History
```
      $ .\migrate.exe -showHistory=true -dsn="root:my-password@tcp(127.0.0.1:46985)/test"
```
This command will return the migration history in the `_MigrationInfo` table.
## Database Schema Files
The database schema files created by this tool follows this format:
```
{version}_{name}_up.sql
{version}_{name}_down.sql
```
Thanks to [golang-migrate/migrate](https://github.com/golang-migrate/migrate) !!!  
The tool uses epoch time as the `{version}`, and takes the `-newSchema` option value as the `{name}`. The `up` files are used for normal database schema migration. The `down` files are used for database schema rollback.