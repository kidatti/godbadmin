package db

import (
	"fmt"
	"godbadmin/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type TableInfo struct {
	TableName string `db:"TABLE_NAME"`
}

type DatabaseInfo struct {
	DatabaseName string `db:"SCHEMA_NAME"`
}

func Connect(server config.ServerConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		server.User,
		server.Password,
		server.Host,
		server.Port,
		server.Database,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectWithoutDB(host string, port int, user, password string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		user,
		password,
		host,
		port,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetDatabases(db *sqlx.DB) ([]DatabaseInfo, error) {
	var databases []DatabaseInfo
	query := `SELECT SCHEMA_NAME FROM information_schema.SCHEMATA
	          WHERE SCHEMA_NAME NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
	          ORDER BY SCHEMA_NAME`

	err := db.Select(&databases, query)
	if err != nil {
		return nil, err
	}

	return databases, nil
}

func GetAllDatabases(db *sqlx.DB) ([]DatabaseInfo, error) {
	var databases []DatabaseInfo
	query := `SELECT SCHEMA_NAME FROM information_schema.SCHEMATA ORDER BY SCHEMA_NAME`

	err := db.Select(&databases, query)
	if err != nil {
		return nil, err
	}

	return databases, nil
}

type DatabaseWithTables struct {
	DatabaseName string
	Tables       []TableInfo
}

func GetTables(db *sqlx.DB, database string) ([]TableInfo, error) {
	var tables []TableInfo
	query := `SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME`

	err := db.Select(&tables, query, database)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func GetTableData(db *sqlx.DB, database, tableName string, limit int) ([]map[string]interface{}, []string, error) {
	if limit <= 0 {
		limit = 100
	}

	// Select database first
	if database != "" {
		_, err := db.Exec(fmt.Sprintf("USE `%s`", database))
		if err != nil {
			return nil, nil, err
		}
	}

	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d", tableName, limit)
	rows, err := db.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return nil, nil, err
		}

		// Convert []byte to string for display
		for key, val := range row {
			if b, ok := val.([]byte); ok {
				row[key] = string(b)
			}
		}

		results = append(results, row)
	}

	return results, columns, nil
}

func GetTableDataAll(db *sqlx.DB, database, tableName string) ([]map[string]interface{}, error) {
	// Select database first
	if database != "" {
		_, err := db.Exec(fmt.Sprintf("USE `%s`", database))
		if err != nil {
			return nil, err
		}
	}

	query := fmt.Sprintf("SELECT * FROM `%s`", tableName)
	rows, err := db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return nil, err
		}

		// Convert []byte to string for display
		for key, val := range row {
			if b, ok := val.([]byte); ok {
				row[key] = string(b)
			}
		}

		results = append(results, row)
	}

	return results, nil
}

type ColumnInfo struct {
	Field   string `db:"Field"`
	Type    string `db:"Type"`
	Null    string `db:"Null"`
	Key     string `db:"Key"`
	Default *string `db:"Default"`
	Extra   string `db:"Extra"`
}

func GetTableColumns(db *sqlx.DB, database, tableName string) ([]ColumnInfo, error) {
	// Select database first
	if database != "" {
		_, err := db.Exec(fmt.Sprintf("USE `%s`", database))
		if err != nil {
			return nil, err
		}
	}

	var columns []ColumnInfo
	query := fmt.Sprintf("SHOW COLUMNS FROM `%s`", tableName)
	err := db.Select(&columns, query)
	if err != nil {
		return nil, err
	}

	return columns, nil
}

func GetTableCreateStatement(db *sqlx.DB, database, tableName string) (string, error) {
	// Select database first
	if database != "" {
		_, err := db.Exec(fmt.Sprintf("USE `%s`", database))
		if err != nil {
			return "", err
		}
	}

	// Use QueryRow and scan into map to handle any number of columns
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)

	var tblName string
	var createStmt string

	err := db.QueryRow(query).Scan(&tblName, &createStmt)
	if err != nil {
		return "", err
	}

	return createStmt, nil
}

func GetPrimaryKeyColumns(db *sqlx.DB, database, tableName string) ([]string, error) {
	// Select database first
	if database != "" {
		_, err := db.Exec(fmt.Sprintf("USE `%s`", database))
		if err != nil {
			return nil, err
		}
	}

	var pkColumns []string
	query := `SELECT COLUMN_NAME
	          FROM information_schema.KEY_COLUMN_USAGE
	          WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'
	          ORDER BY ORDINAL_POSITION`

	err := db.Select(&pkColumns, query, database, tableName)
	if err != nil {
		return nil, err
	}

	return pkColumns, nil
}

func GetRowData(db *sqlx.DB, database, tableName string, pkColumns []string, pkValues []string) (map[string]interface{}, error) {
	// Select database first
	if database != "" {
		_, err := db.Exec(fmt.Sprintf("USE `%s`", database))
		if err != nil {
			return nil, err
		}
	}

	// Build WHERE clause
	var whereClauses []string
	for _, col := range pkColumns {
		whereClauses = append(whereClauses, fmt.Sprintf("`%s` = ?", col))
	}
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + fmt.Sprintf("%s", whereClauses[0])
		for i := 1; i < len(whereClauses); i++ {
			whereClause += " AND " + whereClauses[i]
		}
	}

	query := fmt.Sprintf("SELECT * FROM `%s`%s LIMIT 1", tableName, whereClause)

	// Convert pkValues to interface{} slice
	args := make([]interface{}, len(pkValues))
	for i, v := range pkValues {
		args[i] = v
	}

	rows, err := db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return nil, err
		}

		// Convert []byte to string for display
		for key, val := range row {
			if b, ok := val.([]byte); ok {
				row[key] = string(b)
			}
		}

		return row, nil
	}

	return nil, fmt.Errorf("row not found")
}

type ServerInfo struct {
	Version               string
	VersionComment        string
	ProtocolVersion       string
	ConnectionID          string
	CurrentUser           string
	CharacterSetServer    string
	CollationServer       string
	CharacterSetConnection string
	CollationConnection   string
	SSLCipher             string
}

func GetServerInfo(db *sqlx.DB) (*ServerInfo, error) {
	info := &ServerInfo{}

	// Get version
	var version string
	err := db.Get(&version, "SELECT VERSION()")
	if err == nil {
		info.Version = version
	}

	// Get version comment
	var versionComment string
	err = db.Get(&versionComment, "SELECT @@version_comment")
	if err == nil {
		info.VersionComment = versionComment
	}

	// Get protocol version
	var protocolVersion string
	err = db.Get(&protocolVersion, "SELECT @@protocol_version")
	if err == nil {
		info.ProtocolVersion = protocolVersion
	}

	// Get connection ID
	var connectionID int
	err = db.Get(&connectionID, "SELECT CONNECTION_ID()")
	if err == nil {
		info.ConnectionID = fmt.Sprintf("%d", connectionID)
	}

	// Get current user
	var currentUser string
	err = db.Get(&currentUser, "SELECT CURRENT_USER()")
	if err == nil {
		info.CurrentUser = currentUser
	}

	// Get character set
	var charsetServer string
	err = db.Get(&charsetServer, "SELECT @@character_set_server")
	if err == nil {
		info.CharacterSetServer = charsetServer
	}

	// Get collation
	var collationServer string
	err = db.Get(&collationServer, "SELECT @@collation_server")
	if err == nil {
		info.CollationServer = collationServer
	}

	// Get connection character set
	var charsetConnection string
	err = db.Get(&charsetConnection, "SELECT @@character_set_connection")
	if err == nil {
		info.CharacterSetConnection = charsetConnection
	}

	// Get connection collation
	var collationConnection string
	err = db.Get(&collationConnection, "SELECT @@collation_connection")
	if err == nil {
		info.CollationConnection = collationConnection
	}

	// Get SSL cipher (empty if not using SSL)
	var sslCipher string
	err = db.Get(&sslCipher, "SHOW STATUS LIKE 'Ssl_cipher'")
	if err == nil {
		info.SSLCipher = sslCipher
	}

	return info, nil
}

type UserPrivilege struct {
	User       string
	Host       string
	Privileges string
}

func GetUserPrivileges(db *sqlx.DB) ([]UserPrivilege, error) {
	// Get all users
	type MySQLUser struct {
		User string `db:"User"`
		Host string `db:"Host"`
	}

	var users []MySQLUser
	err := db.Select(&users, "SELECT User, Host FROM mysql.user ORDER BY User, Host")
	if err != nil {
		return nil, err
	}

	var userPrivileges []UserPrivilege
	for _, user := range users {
		// Get grants for each user
		query := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", user.User, user.Host)
		rows, err := db.Query(query)
		if err != nil {
			// Skip users we can't query
			continue
		}

		var grants []string
		for rows.Next() {
			var grant string
			if err := rows.Scan(&grant); err == nil {
				grants = append(grants, grant)
			}
		}
		rows.Close()

		userPrivileges = append(userPrivileges, UserPrivilege{
			User:       user.User,
			Host:       user.Host,
			Privileges: fmt.Sprintf("%d grants", len(grants)),
		})
	}

	return userPrivileges, nil
}

func GetUserGrants(db *sqlx.DB, user, host string) ([]string, error) {
	query := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", user, host)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []string
	for rows.Next() {
		var grant string
		if err := rows.Scan(&grant); err == nil {
			grants = append(grants, grant)
		}
	}

	return grants, nil
}
