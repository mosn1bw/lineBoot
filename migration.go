package main

import (
	"log"

	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func Migrate() {
	log.Printf("Start Migration")

	m := gormigrate.New(repo, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202010071100",
			Migrate: func(tx *gorm.DB) error {
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return nil
			},
		},
	})

	// TODO: add custom type
	m.InitSchema(func(tx *gorm.DB) error {
		log.Printf("Create Tables...")
		if err := repo.AutoMigrate(&Message{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err := m.Migrate(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Migrate Finished")
}
