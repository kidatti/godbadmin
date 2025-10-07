package main

import (
	"flag"
	"fmt"
	"godbadmin/config"
	"godbadmin/handlers"
	"godbadmin/i18n"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func findAvailablePort(startPort int) int {
	port := startPort
	for !isPortAvailable(port) {
		log.Printf("Port %d is already in use, trying %d...", port, port+1)
		port++
		if port > startPort+10 {
			log.Fatal("Could not find available port after 10 attempts")
		}
	}
	return port
}

func main() {
	// Parse command line flags
	portFlag := flag.Int("port", 8000, "Port to run the server on")
	flag.Parse()

	// Load settings
	settings := config.GetSettings()
	if err := settings.Load("settings.json"); err != nil {
		log.Printf("Warning: Could not load settings.json: %v", err)
	}

	// Initialize i18n
	if err := i18n.Init(); err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(i18n.Middleware())

	// Setup template renderer with translation function
	funcMap := template.FuncMap{
		"T": func(c echo.Context, key string) string {
			return i18n.T(c, key)
		},
	}
	renderer := &TemplateRenderer{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(302, "/servers")
	})

	// Server management routes
	e.GET("/servers", handlers.ServersPage)
	e.GET("/servers/new", handlers.AddServerPage)
	e.POST("/servers", handlers.CreateServer)
	e.GET("/servers/:id/edit", handlers.EditServerPage)
	e.GET("/servers/:id/info", handlers.ServerInfoPage)
	e.GET("/servers/:id/privileges", handlers.UserPrivilegesPage)
	e.POST("/servers/:id", handlers.UpdateServer)
	e.POST("/servers/:id/delete", handlers.DeleteServer)

	// API routes
	e.POST("/api/test-connection", handlers.TestConnectionAPI)
	e.GET("/api/databases", handlers.GetDatabasesAPI)
	e.POST("/api/database/create", handlers.CreateDatabaseAPI)
	e.GET("/api/user-grants", handlers.GetUserGrantsAPI)
	e.GET("/api/set-language", func(c echo.Context) error {
		lang := c.QueryParam("lang")
		if lang != "ja" && lang != "en" {
			lang = "ja"
		}

		cookie := &http.Cookie{
			Name:     "lang",
			Value:    lang,
			Path:     "/",
			MaxAge:   365 * 24 * 60 * 60, // 1 year
			HttpOnly: true,
		}
		c.SetCookie(cookie)

		// Redirect to the previous page or home
		referer := c.Request().Header.Get("Referer")
		if referer == "" {
			referer = "/servers"
		}
		return c.Redirect(http.StatusSeeOther, referer)
	})

	// Database routes
	e.GET("/servers/:id/database", handlers.DatabasePage)
	e.GET("/servers/:id/db/:db/table/:table", handlers.TableDataPage)
	e.GET("/servers/:id/db/:db/table/:table/details", handlers.TableDetailsPage)
	e.GET("/servers/:id/db/:db/table/:table/edit", handlers.TableEditPage)
	e.GET("/servers/:id/db/:db/table/:table/delete", handlers.DeleteTable)
	e.GET("/servers/:id/db/:db/table/:table/row", handlers.RowDetailsPage)
	e.GET("/servers/:id/db/:db/export", handlers.ExportPage)
	e.POST("/servers/:id/db/:db/export", handlers.ExportData)
	e.GET("/servers/:id/table/:table", handlers.TableDataPage)   // Legacy route
	e.GET("/servers/:id/tables", handlers.TablesPage)            // Legacy route

	// Find available port
	port := findAvailablePort(*portFlag)
	addr := ":" + strconv.Itoa(port)

	// Start server
	log.Printf("Starting server on http://localhost%s", addr)
	e.Logger.Fatal(e.Start(addr))
}
