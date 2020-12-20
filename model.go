package main

type Message struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	UserID    string `json:"userId,omitempty" gorm:"type:varchar(255);not null"`
	GroupID   string `json:"groupId,omitempty" gorm:"type:varchar(255);not null"`
	RoomID    string `json:"roomId,omitempty" gorm:"type:varchar(255);not null"`
	MessageID string `json:"messageID,omitempty" gorm:"type:varchar(255);not null"`
	Message   string `json:"message,omitempty" gorm:"type:text;not null;default:''"`
}
