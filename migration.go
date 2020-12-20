package main

import (
	"github.com/jinzhu/gorm"
	"github.com/mlytics/go-util/log"
	"gopkg.in/gormigrate.v1"
)

func Migrate() {
	log.Info("Start Migration")

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
		log.Info("Create Tables...")
		if err := repo.AutoMigrate().Error; err != nil {
			return err
		}

		return nil
	})

	if err := m.Migrate(); err != nil {
		log.WithError(err).Error("Database Migration Failed")
	}
	log.Info("Migrate Finished")
}
