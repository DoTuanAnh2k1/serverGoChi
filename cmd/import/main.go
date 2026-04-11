// CLI tool to import users, NEs, user-NE mappings, and user-role mappings
// from a text file into the database.
//
// Usage:
//
//	go run ./cmd/import -file data.txt
//
// File format (sections separated by blank lines, # comments allowed):
//
//	[users]
//	username,password
//	admin,admin123
//	operator1,Pass@123
//
//	[nes]
//	name,site_name,ip_address,port,namespace,description
//	HTSMF01,HCM,10.10.1.1,22,hcm-5gc,HCM SMF Node 01
//
//	[roles]
//	permission,scope,ne_type,include_type,path
//	admin,ext-config,5GC,include,/
//
//	[user_roles]
//	username,permission
//	admin,admin
//	operator1,operator
//
//	[user_nes]
//	username,ne_name
//	admin,HTSMF01
//	operator1,HTAMF01
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/bcrypt"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

func main() {
	filePath := flag.String("file", "", "path to import file")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Usage: go run ./cmd/import -file <path>")
		fmt.Println("\nFile format example:")
		fmt.Println(`
[users]
username,password
admin,admin123

[nes]
name,site_name,ip_address,port,namespace,description
HTSMF01,HCM,10.10.1.1,22,hcm-5gc,HCM SMF Node 01

[roles]
permission,scope,ne_type,include_type,path
admin,ext-config,5GC,include,/

[user_roles]
username,permission
admin,admin

[user_nes]
username,ne_name
admin,HTSMF01`)
		os.Exit(1)
	}

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	initDB()

	sections, err := parseFile(*filePath)
	if err != nil {
		log.Fatalf("parse file: %v", err)
	}

	db := store.GetSingleton()

	// Import users
	if rows, ok := sections["users"]; ok {
		for _, cols := range rows {
			if len(cols) < 2 {
				continue
			}
			username, password := cols[0], cols[1]
			existing, _ := db.GetUserByUserName(username)
			if existing != nil {
				fmt.Printf("  [skip] user %q already exists\n", username)
				continue
			}
			now := time.Now()
			user := &db_models.TblAccount{
				AccountName:   username,
				Password:      bcrypt.Encode(username + password),
				AccountType:   2,
				IsEnable:      true,
				Status:        true,
				CreatedDate:   now,
				UpdatedDate:   now,
				LastLoginTime: now,
				LastChangePass: now,
				LockedTime:    now,
				CreatedBy:     "import",
			}
			if err := db.AddUser(user); err != nil {
				fmt.Printf("  [error] user %q: %v\n", username, err)
			} else {
				fmt.Printf("  [ok] user %q created\n", username)
			}
		}
	}

	// Import NEs
	if rows, ok := sections["nes"]; ok {
		for _, cols := range rows {
			if len(cols) < 6 {
				continue
			}
			port, _ := strconv.Atoi(cols[3])
			ne := &db_models.CliNe{
				Name:        cols[0],
				SiteName:    cols[1],
				IPAddress:   cols[2],
				Port:        int32(port),
				Namespace:   cols[4],
				Description: cols[5],
				SystemType:  "5GC",
			}
			if err := db.CreateCliNe(ne); err != nil {
				fmt.Printf("  [error] ne %q: %v\n", cols[0], err)
			} else {
				fmt.Printf("  [ok] ne %q created (id=%d)\n", cols[0], ne.ID)
			}
		}
	}

	// Import roles
	if rows, ok := sections["roles"]; ok {
		for _, cols := range rows {
			if len(cols) < 5 {
				continue
			}
			role := &db_models.CliRole{
				Permission:  cols[0],
				Scope:       cols[1],
				NeType:      cols[2],
				IncludeType: cols[3],
				Path:        cols[4],
			}
			existing, _ := db.GetCliRole(role)
			if existing != nil {
				fmt.Printf("  [skip] role %q already exists\n", cols[0])
				continue
			}
			if err := db.CreateCliRole(role); err != nil {
				fmt.Printf("  [error] role %q: %v\n", cols[0], err)
			} else {
				fmt.Printf("  [ok] role %q created\n", cols[0])
			}
		}
	}

	// Import user-role mappings
	if rows, ok := sections["user_roles"]; ok {
		for _, cols := range rows {
			if len(cols) < 2 {
				continue
			}
			username, permission := cols[0], cols[1]
			user, err := db.GetUserByUserName(username)
			if err != nil || user == nil {
				fmt.Printf("  [error] user_role: user %q not found\n", username)
				continue
			}
			if err := db.AddRole(&db_models.CliRoleUserMapping{UserID: user.AccountID, Permission: permission}); err != nil {
				fmt.Printf("  [error] user_role %q->%q: %v\n", username, permission, err)
			} else {
				fmt.Printf("  [ok] user_role %q -> %q\n", username, permission)
			}
		}
	}

	// Import user-NE mappings
	if rows, ok := sections["user_nes"]; ok {
		// build NE name -> ID lookup
		neMap := map[string]int64{}
		allNes, _ := db.GetCliNeListBySystemType("5GC")
		for _, ne := range allNes {
			neMap[ne.Name] = ne.ID
		}

		for _, cols := range rows {
			if len(cols) < 2 {
				continue
			}
			username, neName := cols[0], cols[1]
			user, err := db.GetUserByUserName(username)
			if err != nil || user == nil {
				fmt.Printf("  [error] user_ne: user %q not found\n", username)
				continue
			}
			neID, ok := neMap[neName]
			if !ok {
				fmt.Printf("  [error] user_ne: ne %q not found\n", neName)
				continue
			}
			if err := db.CreateUserNeMapping(&db_models.CliUserNeMapping{UserID: user.AccountID, TblNeID: neID}); err != nil {
				fmt.Printf("  [error] user_ne %q->%q: %v\n", username, neName, err)
			} else {
				fmt.Printf("  [ok] user_ne %q -> %q\n", username, neName)
			}
		}
	}

	fmt.Println("\nImport complete.")
}

func initDB() {
	cfg := &config_models.Config{
		Db: config_models.DatabaseConfig{
			DbType: os.Getenv("DB_DRIVER"),
			Mysql: config_models.MySqlConfig{
				Host:     os.Getenv("MYSQL_HOST"),
				Port:     os.Getenv("MYSQL_PORT"),
				User:     os.Getenv("MYSQL_USER"),
				Password: os.Getenv("MYSQL_PASSWORD"),
				Name:     os.Getenv("MYSQL_DB_NAME"),
			},
			Postgres: config_models.PostgresConfig{
				Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
				Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
				User:     os.Getenv("POSTGRES_USER"),
				Password: os.Getenv("POSTGRES_PASSWORD"),
				Name:     getEnvOrDefault("POSTGRES_DB_NAME", "cli_db"),
				SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
			},
			Mongo: config_models.MongoConfig{
				URI:      getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
				Database: getEnvOrDefault("MONGODB_DB_NAME", "cli_db"),
			},
		},
		Log: config_models.LogConfig{
			Level:   getEnvOrDefault("LOG_LEVEL", "error"),
			DbLevel: getEnvOrDefault("DB_LOG_LEVEL", "error"),
		},
	}
	config.Init(cfg)
	logger.Init(cfg.Log.Level, cfg.Log.DbLevel)
	store.Init()
}

// parseFile reads a sectioned text file and returns map[section][]row.
// Each row is a slice of trimmed CSV fields.
func parseFile(path string) (map[string][][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sections := map[string][][]string{}
	var current string
	headerSkipped := map[string]bool{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.ToLower(line[1 : len(line)-1])
			continue
		}
		if current == "" {
			continue
		}
		// skip header row (first row after section tag)
		if !headerSkipped[current] {
			headerSkipped[current] = true
			continue
		}
		cols := strings.Split(line, ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}
		sections[current] = append(sections[current], cols)
	}
	return sections, scanner.Err()
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
