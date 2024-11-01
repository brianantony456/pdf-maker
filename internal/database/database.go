package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	// Driver for sqlite3
	_ "github.com/mattn/go-sqlite3"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	GetColumns() ([]map[string]interface{}, error)
	AddColumn(name, dataType string) error
	AddRow(data map[string]string) error
	GetRow(id int) (map[string]string, error)
	UpdateRow(id int, data map[string]string) error
	GetAllRows() ([]map[string]interface{}, error)
}

type serviceImpl struct {
	db *sql.DB
}

func createInitialTables(s *serviceImpl) {
	dataQuery := `
	CREATE TABLE IF NOT EXISTS data_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT
	);
	`
	_, err := s.db.Exec(dataQuery)
	if err != nil {
		log.Fatalf("Failed to create data_entries table: %v", err)
	}

	logQuery := `
	CREATE TABLE IF NOT EXISTS update_log (
		id INTEGER,
		timestamp TEXT,
		data TEXT
	);
	`
	_, err = s.db.Exec(logQuery)
	if err != nil {
		log.Fatalf("Failed to create update_log table: %v", err)
	}
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *serviceImpl) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(stats["error"]) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *serviceImpl) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	return s.db.Close()
}

var dburl = os.Getenv("BLUEPRINT_DB_URL")
var dbInstance *serviceImpl

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}

	dbInstance = &serviceImpl{
		db: db,
	}

	createInitialTables(dbInstance)
	return dbInstance
}
