package handlers

import (
	"encoding/csv"
	"fmt"
	"godbadmin/config"
	"godbadmin/db"
	"godbadmin/i18n"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Helper function to add Context and Lang to template data
func addI18nContext(c echo.Context, data map[string]interface{}) map[string]interface{} {
	data["Context"] = c
	data["Lang"] = i18n.GetCurrentLang(c)
	return data
}

func DatabasePage(c echo.Context) error {
	serverID := c.Param("id")
	dbName := c.QueryParam("db")
	settings := config.GetSettings()

	server, found := settings.GetServer(serverID)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	// If no database specified, use the default one from server config
	if dbName == "" {
		dbName = server.Database
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Render(http.StatusOK, "database_overview.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース接続エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        "",
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  true,
		}))
	}
	defer dbConn.Close()

	// Get all databases
	databases, err := db.GetAllDatabases(dbConn)
	if err != nil {
		return c.Render(http.StatusOK, "database_overview.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース一覧の取得エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        "",
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  true,
		}))
	}

	// Get tables for each database
	var dbWithTables []db.DatabaseWithTables
	for _, database := range databases {
		tables, _ := db.GetTables(dbConn, database.DatabaseName)
		dbWithTables = append(dbWithTables, db.DatabaseWithTables{
			DatabaseName: database.DatabaseName,
			Tables:       tables,
		})
	}

	// Get tables for current database
	currentTables, _ := db.GetTables(dbConn, dbName)

	return c.Render(http.StatusOK, "database_overview.html", addI18nContext(c, map[string]interface{}{
		"Server":              server,
		"Error":               "",
		"DatabasesWithTables": dbWithTables,
		"CurrentDatabase":     dbName,
		"CurrentTable":        "",
		"Tables":              currentTables,
		"ActiveMenu":          "database",
		"ShowDatabaseDropdown": true,
		"ShowCreateDatabase":  true,
	}))
}

func TableDataPage(c echo.Context) error {
	serverID := c.Param("id")
	dbName := c.Param("db")
	tableName := c.Param("table")
	settings := config.GetSettings()

	server, found := settings.GetServer(serverID)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	// If db is not provided in URL, use server's default database
	if dbName == "" {
		dbName = server.Database
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Render(http.StatusOK, "table_data.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース接続エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
		}))
	}
	defer dbConn.Close()

	// Get all databases and their tables
	databases, err := db.GetAllDatabases(dbConn)
	if err != nil {
		return c.Render(http.StatusOK, "table_data.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース一覧の取得エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
		}))
	}

	var dbWithTables []db.DatabaseWithTables
	for _, database := range databases {
		tables, _ := db.GetTables(dbConn, database.DatabaseName)
		dbWithTables = append(dbWithTables, db.DatabaseWithTables{
			DatabaseName: database.DatabaseName,
			Tables:       tables,
		})
	}

	// Get table data
	tableData, columns, err := db.GetTableData(dbConn, dbName, tableName, 100)
	if err != nil {
		return c.Render(http.StatusOK, "table_data.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "テーブルデータの取得エラー: " + err.Error(),
			"DatabasesWithTables": dbWithTables,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"TableData":           nil,
			"Columns":             nil,
			"PrimaryKeys":         nil,
		}))
	}

	// Get primary key columns
	pkColumns, _ := db.GetPrimaryKeyColumns(dbConn, dbName, tableName)

	return c.Render(http.StatusOK, "table_data.html", addI18nContext(c, map[string]interface{}{
		"Server":              server,
		"Error":               "",
		"DatabasesWithTables": dbWithTables,
		"CurrentDatabase":     dbName,
		"CurrentTable":        tableName,
		"TableData":           tableData,
		"Columns":             columns,
		"PrimaryKeys":         pkColumns,
	}))
}

// Legacy function for backward compatibility
func TablesPage(c echo.Context) error {
	return DatabasePage(c)
}

type CreateDatabaseRequest struct {
	ServerID  string `json:"server_id"`
	DBName    string `json:"db_name"`
	Charset   string `json:"charset"`
	Collation string `json:"collation"`
}

func CreateDatabaseAPI(c echo.Context) error {
	var req CreateDatabaseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "Invalid request",
		})
	}

	settings := config.GetSettings()
	server, found := settings.GetServer(req.ServerID)
	if !found {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "Server not found",
		})
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "Database connection error: " + err.Error(),
		})
	}
	defer dbConn.Close()

	// Build CREATE DATABASE query
	query := fmt.Sprintf("CREATE DATABASE `%s`", req.DBName)
	if req.Charset != "" {
		query += fmt.Sprintf(" CHARACTER SET %s", req.Charset)
	}
	if req.Collation != "" {
		query += fmt.Sprintf(" COLLATE %s", req.Collation)
	}

	_, err = dbConn.Exec(query)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "Failed to create database: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Database created successfully",
	})
}

func TableDetailsPage(c echo.Context) error {
	serverID := c.Param("id")
	dbName := c.Param("db")
	tableName := c.Param("table")
	settings := config.GetSettings()

	server, found := settings.GetServer(serverID)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	// If db is not provided in URL, use server's default database
	if dbName == "" {
		dbName = server.Database
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Render(http.StatusOK, "table_details.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース接続エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}
	defer dbConn.Close()

	// Get all databases and their tables
	databases, err := db.GetAllDatabases(dbConn)
	if err != nil {
		return c.Render(http.StatusOK, "table_details.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース一覧の取得エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	var dbWithTables []db.DatabaseWithTables
	for _, database := range databases {
		tables, _ := db.GetTables(dbConn, database.DatabaseName)
		dbWithTables = append(dbWithTables, db.DatabaseWithTables{
			DatabaseName: database.DatabaseName,
			Tables:       tables,
		})
	}

	// Get table columns
	columns, err := db.GetTableColumns(dbConn, dbName, tableName)
	if err != nil {
		return c.Render(http.StatusOK, "table_details.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "カラム情報の取得エラー: " + err.Error(),
			"DatabasesWithTables": dbWithTables,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"Columns":             nil,
			"CreateStatement":     "",
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	// Get CREATE TABLE statement
	createStmt, err := db.GetTableCreateStatement(dbConn, dbName, tableName)
	if err != nil {
		return c.Render(http.StatusOK, "table_details.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "CREATE TABLE文の取得エラー: " + err.Error(),
			"DatabasesWithTables": dbWithTables,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"Columns":             columns,
			"CreateStatement":     "",
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	return c.Render(http.StatusOK, "table_details.html", addI18nContext(c, map[string]interface{}{
		"Server":              server,
		"Error":               "",
		"DatabasesWithTables": dbWithTables,
		"CurrentDatabase":     dbName,
		"CurrentTable":        tableName,
		"Columns":             columns,
		"CreateStatement":     createStmt,
		"ActiveMenu":          "database",
		"ShowDatabaseDropdown": true,
		"ShowCreateDatabase":  false,
	}))
}

func RowDetailsPage(c echo.Context) error {
	serverID := c.Param("id")
	dbName := c.Param("db")
	tableName := c.Param("table")
	settings := config.GetSettings()

	server, found := settings.GetServer(serverID)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	// If db is not provided in URL, use server's default database
	if dbName == "" {
		dbName = server.Database
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Render(http.StatusOK, "row_details.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース接続エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}
	defer dbConn.Close()

	// Get all databases and their tables
	databases, err := db.GetAllDatabases(dbConn)
	if err != nil {
		return c.Render(http.StatusOK, "row_details.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース一覧の取得エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	var dbWithTables []db.DatabaseWithTables
	for _, database := range databases {
		tables, _ := db.GetTables(dbConn, database.DatabaseName)
		dbWithTables = append(dbWithTables, db.DatabaseWithTables{
			DatabaseName: database.DatabaseName,
			Tables:       tables,
		})
	}

	// Get primary key columns
	pkColumns, err := db.GetPrimaryKeyColumns(dbConn, dbName, tableName)
	if err != nil || len(pkColumns) == 0 {
		return c.Render(http.StatusOK, "row_details.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "主キーが見つかりません",
			"DatabasesWithTables": dbWithTables,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	// Get primary key values from query parameters
	pkValues := make([]string, len(pkColumns))
	for i, col := range pkColumns {
		pkValues[i] = c.QueryParam(col)
		if pkValues[i] == "" {
			return c.Render(http.StatusOK, "row_details.html", addI18nContext(c, map[string]interface{}{
				"Server":              server,
				"Error":               "主キー値が指定されていません",
				"DatabasesWithTables": dbWithTables,
				"CurrentDatabase":     dbName,
				"CurrentTable":        tableName,
				"ActiveMenu":          "database",
				"ShowDatabaseDropdown": true,
				"ShowCreateDatabase":  false,
			}))
		}
	}

	// Get row data
	rowData, err := db.GetRowData(dbConn, dbName, tableName, pkColumns, pkValues)
	if err != nil {
		return c.Render(http.StatusOK, "row_details.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "行データの取得エラー: " + err.Error(),
			"DatabasesWithTables": dbWithTables,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	// Get column information
	columns, _ := db.GetTableColumns(dbConn, dbName, tableName)

	return c.Render(http.StatusOK, "row_details.html", addI18nContext(c, map[string]interface{}{
		"Server":              server,
		"Error":               "",
		"DatabasesWithTables": dbWithTables,
		"CurrentDatabase":     dbName,
		"CurrentTable":        tableName,
		"RowData":             rowData,
		"Columns":             columns,
		"ActiveMenu":          "database",
		"ShowDatabaseDropdown": true,
		"ShowCreateDatabase":  false,
	}))
}

func ExportPage(c echo.Context) error {
	serverID := c.Param("id")
	dbName := c.Param("db")
	selectedTable := c.QueryParam("table")
	settings := config.GetSettings()

	server, found := settings.GetServer(serverID)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Render(http.StatusOK, "export.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース接続エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"Tables":              nil,
			"SelectedTable":       selectedTable,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}
	defer dbConn.Close()

	// Get all databases and their tables
	databases, err := db.GetAllDatabases(dbConn)
	if err != nil {
		return c.Render(http.StatusOK, "export.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース一覧の取得エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"Tables":              nil,
			"SelectedTable":       selectedTable,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	var dbWithTables []db.DatabaseWithTables
	for _, database := range databases {
		tables, _ := db.GetTables(dbConn, database.DatabaseName)
		dbWithTables = append(dbWithTables, db.DatabaseWithTables{
			DatabaseName: database.DatabaseName,
			Tables:       tables,
		})
	}

	// Get tables for current database
	tables, err := db.GetTables(dbConn, dbName)
	if err != nil {
		return c.Render(http.StatusOK, "export.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "テーブル一覧の取得エラー: " + err.Error(),
			"DatabasesWithTables": dbWithTables,
			"CurrentDatabase":     dbName,
			"Tables":              nil,
			"SelectedTable":       selectedTable,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	return c.Render(http.StatusOK, "export.html", addI18nContext(c, map[string]interface{}{
		"Server":              server,
		"Error":               "",
		"DatabasesWithTables": dbWithTables,
		"CurrentDatabase":     dbName,
		"Tables":              tables,
		"SelectedTable":       selectedTable,
		"ActiveMenu":          "database",
		"ShowDatabaseDropdown": true,
		"ShowCreateDatabase":  false,
	}))
}

func ExportData(c echo.Context) error {
	serverID := c.Param("id")
	dbName := c.Param("db")
	settings := config.GetSettings()

	server, found := settings.GetServer(serverID)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "フォームデータの解析エラー")
	}

	// Get form data
	tables := c.Request().Form["tables"]
	_ = c.FormValue("format") // Currently only CSV is supported
	delimiter := c.FormValue("delimiter")
	includeHeaders := c.FormValue("include_headers") == "true"
	encoding := c.FormValue("encoding")

	if len(tables) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "テーブルが選択されていません")
	}

	// Convert delimiter
	if delimiter == "\\t" {
		delimiter = "\t"
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "データベース接続エラー: "+err.Error())
	}
	defer dbConn.Close()

	// Set response headers
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_export.csv", dbName))
	c.Response().WriteHeader(http.StatusOK)

	// Create CSV writer
	writer := csv.NewWriter(c.Response().Writer)
	writer.Comma = rune(delimiter[0])

	// Handle encoding
	var w *csv.Writer
	if encoding == "sjis" {
		encoder := japanese.ShiftJIS.NewEncoder()
		w = csv.NewWriter(transform.NewWriter(c.Response().Writer, encoder))
		w.Comma = rune(delimiter[0])
	} else if encoding == "eucjp" {
		encoder := japanese.EUCJP.NewEncoder()
		w = csv.NewWriter(transform.NewWriter(c.Response().Writer, encoder))
		w.Comma = rune(delimiter[0])
	} else {
		w = writer
	}

	// Export each table
	for tableIndex, tableName := range tables {
		// Get table data
		data, err := db.GetTableDataAll(dbConn, dbName, tableName)
		if err != nil {
			continue
		}

		if len(data) == 0 {
			continue
		}

		// Get column names
		columns, err := db.GetTableColumns(dbConn, dbName, tableName)
		if err != nil {
			continue
		}

		var columnNames []string
		for _, col := range columns {
			columnNames = append(columnNames, col.Field)
		}

		// Write headers
		if includeHeaders {
			w.Write(columnNames)
		}

		// Write data
		for _, row := range data {
			var record []string
			for _, colName := range columnNames {
				val := row[colName]
				valStr := fmt.Sprintf("%v", val)
				record = append(record, valStr)
			}
			w.Write(record)
		}

		// Add blank line between tables (except for the last table)
		if tableIndex < len(tables)-1 {
			w.Write([]string{})
		}
	}

	w.Flush()
	return nil
}

// TableEditPage shows the table edit page
func TableEditPage(c echo.Context) error {
	serverID := c.Param("id")
	dbName := c.Param("db")
	tableName := c.Param("table")
	settings := config.GetSettings()

	server, found := settings.GetServer(serverID)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Render(http.StatusOK, "table_edit.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース接続エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"Columns":             nil,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}
	defer dbConn.Close()

	// Get all databases and their tables for the sidebar
	databases, err := db.GetAllDatabases(dbConn)
	if err != nil {
		return c.Render(http.StatusOK, "table_edit.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "データベース一覧の取得エラー: " + err.Error(),
			"DatabasesWithTables": nil,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"Columns":             nil,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	var dbWithTables []db.DatabaseWithTables
	for _, database := range databases {
		tables, _ := db.GetTables(dbConn, database.DatabaseName)
		dbWithTables = append(dbWithTables, db.DatabaseWithTables{
			DatabaseName: database.DatabaseName,
			Tables:       tables,
		})
	}

	// Get column information
	columns, err := db.GetTableColumns(dbConn, dbName, tableName)
	if err != nil {
		return c.Render(http.StatusOK, "table_edit.html", addI18nContext(c, map[string]interface{}{
			"Server":              server,
			"Error":               "カラム情報の取得エラー: " + err.Error(),
			"DatabasesWithTables": dbWithTables,
			"CurrentDatabase":     dbName,
			"CurrentTable":        tableName,
			"Columns":             nil,
			"ActiveMenu":          "database",
			"ShowDatabaseDropdown": true,
			"ShowCreateDatabase":  false,
		}))
	}

	return c.Render(http.StatusOK, "table_edit.html", addI18nContext(c, map[string]interface{}{
		"Server":              server,
		"Error":               "",
		"DatabasesWithTables": dbWithTables,
		"CurrentDatabase":     dbName,
		"CurrentTable":        tableName,
		"Columns":             columns,
		"ActiveMenu":          "database",
		"ShowDatabaseDropdown": true,
		"ShowCreateDatabase":  false,
	}))
}

// DeleteTable handles table deletion
func DeleteTable(c echo.Context) error {
	serverID := c.Param("id")
	dbName := c.Param("db")
	tableName := c.Param("table")
	settings := config.GetSettings()

	server, found := settings.GetServer(serverID)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/servers/%s/database?db=%s&error=%s", serverID, dbName, "データベース接続エラー"))
	}
	defer dbConn.Close()

	// Execute DROP TABLE
	query := fmt.Sprintf("DROP TABLE `%s`.`%s`", dbName, tableName)
	_, err = dbConn.Exec(query)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/servers/%s/database?db=%s&error=%s", serverID, dbName, "テーブル削除エラー: "+err.Error()))
	}

	// Redirect back to database overview
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/servers/%s/database?db=%s", serverID, dbName))
}
