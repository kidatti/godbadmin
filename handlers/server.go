package handlers

import (
	"godbadmin/config"
	"godbadmin/db"
	"godbadmin/i18n"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	i18nlib "github.com/nicksnyder/go-i18n/v2/i18n"
)

const settingsFile = "settings.json"

func TestConnectionAPI(c echo.Context) error {
	var req struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBType   string `json:"db_type"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
	}

	// Try to connect to the database
	dbConn, err := db.ConnectWithoutDB(req.Host, req.Port, req.User, req.Password)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	defer dbConn.Close()

	// Test the connection by pinging
	if err := dbConn.Ping(); err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func GetDatabasesAPI(c echo.Context) error {
	host := c.QueryParam("host")
	portStr := c.QueryParam("port")
	user := c.QueryParam("user")
	password := c.QueryParam("password")

	port := 3306
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	dbConn, err := db.ConnectWithoutDB(host, port, user, password)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"databases": []string{},
		})
	}
	defer dbConn.Close()

	databases, err := db.GetDatabases(dbConn)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"error":     err.Error(),
			"databases": []string{},
		})
	}

	dbNames := make([]string, len(databases))
	for i, db := range databases {
		dbNames[i] = db.DatabaseName
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"databases": dbNames,
	})
}

func ServersPage(c echo.Context) error {
	settings := config.GetSettings()
	servers := settings.GetServers()
	selectedID := c.QueryParam("selected")

	var selectedServer *config.ServerConfig
	var databases []db.DatabaseInfo
	var errorMsg string

	// If a server is selected, get its databases
	if selectedID != "" {
		server, found := settings.GetServer(selectedID)
		if found {
			selectedServer = server

			// Try to connect and get databases
			dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
			if err != nil {
				errorMsg = "データベース接続エラー: " + err.Error()
			} else {
				databases, err = db.GetAllDatabases(dbConn)
				if err != nil {
					errorMsg = "データベース一覧の取得エラー: " + err.Error()
				}
				dbConn.Close()
			}
		}
	}

	return c.Render(http.StatusOK, "servers.html", map[string]interface{}{
		"Servers":        servers,
		"SelectedServer": selectedServer,
		"Databases":      databases,
		"Error":          errorMsg,
		"Context":        c,
		"Lang":           i18n.GetCurrentLang(c),
		"ActiveMenu":     "servers",
	})
}

func AddServerPage(c echo.Context) error {
	settings := config.GetSettings()
	servers := settings.GetServers()

	localizer := c.Get("localizer").(*i18nlib.Localizer)
	return c.Render(http.StatusOK, "server_form.html", map[string]interface{}{
		"Title":      localizer.MustLocalize(&i18nlib.LocalizeConfig{MessageID: "add_server_title"}),
		"Action":     "/servers",
		"Server":     nil,
		"Servers":    servers,
		"Databases":  nil,
		"Context":    c,
		"Lang":       i18n.GetCurrentLang(c),
		"ActiveMenu": "servers",
	})
}

func EditServerPage(c echo.Context) error {
	id := c.Param("id")
	settings := config.GetSettings()

	server, found := settings.GetServer(id)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	servers := settings.GetServers()

	localizer := c.Get("localizer").(*i18nlib.Localizer)
	return c.Render(http.StatusOK, "server_form.html", map[string]interface{}{
		"Title":      localizer.MustLocalize(&i18nlib.LocalizeConfig{MessageID: "edit_server_title"}),
		"Action":     "/servers/" + id,
		"Server":     server,
		"Servers":    servers,
		"Context":    c,
		"Lang":       i18n.GetCurrentLang(c),
		"ActiveMenu": "servers",
	})
}

func CreateServer(c echo.Context) error {
	settings := config.GetSettings()

	server := config.ServerConfig{
		ID:       uuid.New().String(),
		Name:     c.FormValue("name"),
		DBType:   c.FormValue("db_type"),
		Host:     c.FormValue("host"),
		User:     c.FormValue("user"),
		Password: c.FormValue("password"),
		Database: "", // No longer use database field from form
	}

	// Parse port
	port := 3306
	if portStr := c.FormValue("port"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	server.Port = port

	settings.AddServer(server)
	if err := settings.Save(settingsFile); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/servers")
}

func UpdateServer(c echo.Context) error {
	id := c.Param("id")
	settings := config.GetSettings()

	server := config.ServerConfig{
		ID:       id,
		Name:     c.FormValue("name"),
		DBType:   c.FormValue("db_type"),
		Host:     c.FormValue("host"),
		User:     c.FormValue("user"),
		Password: c.FormValue("password"),
		Database: "", // No longer use database field from form
	}

	port := 3306
	if portStr := c.FormValue("port"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	server.Port = port

	if !settings.UpdateServer(id, server) {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	if err := settings.Save(settingsFile); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/servers")
}

func DeleteServer(c echo.Context) error {
	id := c.Param("id")
	settings := config.GetSettings()

	if !settings.DeleteServer(id) {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	if err := settings.Save(settingsFile); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/servers")
}

func ServerInfoPage(c echo.Context) error {
	id := c.Param("id")
	settings := config.GetSettings()

	server, found := settings.GetServer(id)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Render(http.StatusOK, "server_info.html", map[string]interface{}{
			"Server":     server,
			"Error":      "データベース接続エラー: " + err.Error(),
			"ServerInfo": nil,
			"ActiveMenu": "servers",
			"Context":    c,
			"Lang":       i18n.GetCurrentLang(c),
		})
	}
	defer dbConn.Close()

	serverInfo, err := db.GetServerInfo(dbConn)
	if err != nil {
		return c.Render(http.StatusOK, "server_info.html", map[string]interface{}{
			"Server":     server,
			"Error":      "サーバ情報の取得エラー: " + err.Error(),
			"ServerInfo": nil,
			"ActiveMenu": "servers",
			"Context":    c,
			"Lang":       i18n.GetCurrentLang(c),
		})
	}

	return c.Render(http.StatusOK, "server_info.html", map[string]interface{}{
		"Server":     server,
		"Error":      "",
		"ServerInfo": serverInfo,
		"ActiveMenu": "servers",
		"Context":    c,
		"Lang":       i18n.GetCurrentLang(c),
	})
}

func UserPrivilegesPage(c echo.Context) error {
	id := c.Param("id")
	settings := config.GetSettings()

	server, found := settings.GetServer(id)
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "Server not found")
	}

	dbConn, err := db.ConnectWithoutDB(server.Host, server.Port, server.User, server.Password)
	if err != nil {
		return c.Render(http.StatusOK, "user_privileges.html", map[string]interface{}{
			"Server":         server,
			"Error":          "データベース接続エラー: " + err.Error(),
			"UserPrivileges": nil,
			"ActiveMenu":     "servers",
			"Context":        c,
			"Lang":           i18n.GetCurrentLang(c),
		})
	}
	defer dbConn.Close()

	userPrivileges, err := db.GetUserPrivileges(dbConn)
	if err != nil {
		return c.Render(http.StatusOK, "user_privileges.html", map[string]interface{}{
			"Server":         server,
			"Error":          "ユーザー権限の取得エラー: " + err.Error(),
			"UserPrivileges": nil,
			"ActiveMenu":     "servers",
			"Context":        c,
			"Lang":           i18n.GetCurrentLang(c),
		})
	}

	return c.Render(http.StatusOK, "user_privileges.html", map[string]interface{}{
		"Server":         server,
		"Error":          "",
		"UserPrivileges": userPrivileges,
		"ActiveMenu":     "servers",
		"Context":        c,
		"Lang":           i18n.GetCurrentLang(c),
	})
}

func GetUserGrantsAPI(c echo.Context) error {
	serverID := c.QueryParam("server_id")
	user := c.QueryParam("user")
	host := c.QueryParam("host")

	if serverID == "" || user == "" || host == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "Missing parameters",
		})
	}

	settings := config.GetSettings()
	server, found := settings.GetServer(serverID)
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

	grants, err := db.GetUserGrants(dbConn, user, host)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "Failed to get grants: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"grants":  grants,
	})
}
