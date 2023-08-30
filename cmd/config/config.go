package config

import (
	"os"
	"fmt"
)

func ConfigSetup() string {
	// Database settings
	os.Setenv("DB_USERNAME", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_NAME", "segments")
	
	dbconfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
							os.Getenv("DB_HOST"), 
							os.Getenv("DB_PORT"), 
							os.Getenv("DB_USERNAME"), 
							os.Getenv("DB_PASSWORD"), 
							os.Getenv("DB_NAME"),
						)

	return dbconfig
}