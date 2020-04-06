package lib

type DbQuery struct{}

// PrepareInitQuery generates the sql script to create migration information table.
func (query *DbQuery) PrepareInitQuery(tableName string) string {
	if tableName == "" {
		return ""
	}
	initQuery := `create table if not exists ` + tableName + ` (
		version varchar(200) null default null
	);`
	return initQuery
}

// PrepareCheckVersionQuery generates the sql script to check existing migration
// version in the target database.
func (query *DbQuery) PrepareCheckVersionQuery(tableName string) string {
	if tableName == "" {
		return ""
	}
	sqlQuery := `select version from ` + tableName + `;`
	return sqlQuery
}

func (query *DbQuery) PrepareLatestVersionQuery(tableName string) string {
	if tableName == "" {
		return ""
	}
	sqlQuery := `select version from ` + tableName + ` order by version desc limit 1;`
	return sqlQuery
}

// PrepareInsertVersionQuery generates the sql script to store migration information.
func (query *DbQuery) PrepareInsertVersionQuery(tableName string) string {
	sqlQuery := `insert into ` + tableName + `(version) values(?);`
	return sqlQuery
}

// PrepareDeleteVersionQuery generates the sql script to delete migration information.
func (query *DbQuery) PrepareDeleteVersionQuery(tableName string) string {
	sqlQuery := `delete from ` + tableName + ` where version=?`
	return sqlQuery
}
