package cmd

import (
	"fmt"
	"log"
	"os"

	"ADRIFT-backend/database/mappers"
	"ADRIFT-backend/database/migrations"
	"ADRIFT-backend/database/seeders"

	"gorm.io/gorm"
)

func Command(db *gorm.DB) {
	migrate := false
	seed := false
	frs := false
	help := false

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--migrate":
			migrate = true
		case "--seed":
			seed = true
		case "--frs":
			frs = true
		case "--help":
			help = true
		}
	}

	if migrate {
		log.Println("Running migration...")
		if err := migrations.Migrate(db); err != nil {
			log.Fatalf("Error migration: %v", err)
		}
		log.Println("✅ Migration completed successfully.")
	}

	if seed {
		log.Println("Running seeder...")
		if err := seeders.Seeder(db); err != nil {
			log.Fatalf("Error seeding: %v", err)
		}
		log.Println("✅ Seeding completed successfully.")
	}

	if frs {
		log.Println("Running FRS mapper...")
		mappers.MapperFRS()
		log.Println("✅ FRS mapping completed successfully.")
	}

	if help {
		fmt.Println(`
		Boilerplate Backend - CLI Commands

		Usage:
		go run main.go [command]

		Commands:
			--migrate    Run database migrations
			--seed       Run database seeders
			--frs        Run FRS Excel mapper
			--help       Show this help message

		Examples:
			go run main.go --migrate
			go run main.go --seed
			go run main.go --frs
			go run main.go --help
		`)
	}
}
