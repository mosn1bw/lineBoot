package main

import (
	"github.com/jinzhu/gorm"
)

// ListMessages list Messages
func ListMessages() (Messages []Message, err error) {
	if err = repo.Find(&Messages).Error; err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	return Messages, nil
}

// CreateMessage create Message
func CreateMessage(m Message) (Message, error) {
	if err := repo.Create(&m).Error; err != nil {
		return m, err
	}

	return m, nil
}
