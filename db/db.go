package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vynious/ascenda-lp-backend/types"
	"github.com/vynious/ascenda-lp-backend/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type IDBService interface {
	CreateTransaction(ctx context.Context, action types.MakerAction, makerId, description string) (*types.Transaction, error)
	UpdateTransaction(ctx context.Context, txnId string, checkerId string, approval bool) (*types.Transaction, error)
	GetCheckers(ctx context.Context, makerId string, role string) ([]string, error)
	GetPoints(ctx context.Context, userId string) ([]string, error)
}

type DB struct {
	Conn    *gorm.DB
	timeout time.Duration
}

type DBService struct {
	ConnMap map[string]*DB
	timeout time.Duration
}

func CreateDBIfNotExists(bank string) error {
	conn, err := connectToDB("ascenda")
	db, _ := conn.DB()
	defer db.Close()

	if err != nil {
		log.Printf("Failed to connect to default ascenda db")
	}

	if err := conn.Exec(fmt.Sprintf("CREATE DATABASE %s", bank)).Error; err != nil {
		log.Printf("failed to create database %s", err)
	}
	return nil
}

func GetAvailableDatabases() []string {
	conn, err := connectToDB("ascenda")
	db, _ := conn.DB()
	defer db.Close()
	if err != nil {
		log.Printf("Failed to connect to default ascenda db")
	}

	log.Printf("GetAvailableDatabases")
	var databases []string
	if err := conn.Raw("SELECT datname FROM pg_database WHERE datistemplate = false and datname not in ('rdsadmin', 'postgres');").Pluck("datname", &databases).Error; err != nil {
		log.Fatalf("Failed to fetch databases: %v", err)
	}

	log.Printf("Fetch available databases %s", databases)
	return databases
}

func SpawnDBService() (*DBService, error) {
	dbService := &DBService{
		ConnMap: make(map[string]*DB),
	}

	// Connect to banks and populate the ConnMap
	bankDBs := GetAvailableDatabases()

	// bankDBs := []string{"ascenda", "bankb", "bankz"} // Add more banks as needed
	for _, bank := range bankDBs {
		conn, err := connectToDB(bank)
		if err != nil {
			log.Fatalf("failed to connect to %s: %v", bank, err)
		}
		dbService.ConnMap[bank] = &DB{Conn: conn}
	}

	log.Println("Successfully connected to all databases")
	return dbService, nil
}

func connectToDB(bank string) (*gorm.DB, error) {
	dbUser := os.Getenv("dbUser")
	// dbName := os.Getenv("dbName")
	dbHost := os.Getenv("dbHost")
	dbPwd := os.Getenv("dbPwd")
	dbName := bank
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		dbHost, 5432, dbUser, dbPwd, dbName,
	)
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to make connection")
	}
	log.Printf("Successfully connected to Database %s", conn.Name())

	return conn, nil
}

func (DBService *DBService) GetBanksDB(token string) *DB {
	bank, err := util.GetCustomAttributeWithCognito("custom:bank", token)
	if err != nil {
		log.Printf("error decoding token to get custom:bank attribute")
	}
	DB := DBService.ConnMap[bank]

	return DB
}

// CloseConn closes connection to db
func (DBService *DBService) CloseConnections() error {
	for _, conn := range DBService.ConnMap {
		db, _ := conn.Conn.DB()
		if err := db.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %v", err)
		}
	}
	return nil
}
