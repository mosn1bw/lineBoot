package main

type Message struct {
	ID            uint   `json:"id" gorm:"primary_key"`
	UserID        string `json:"userId,omitempty" gorm:"type:varchar(255);not null"`
	GroupID       string `json:"groupId,omitempty" gorm:"type:varchar(255);not null"`
	RoomID        string `json:"roomId,omitempty" gorm:"type:varchar(255);not null"`
	SenderName    string `json:"senderName,omitempty" gorm:"type:varchar(255);not null"`
	SenderIconURL string `json:"senderIconURL,omitempty" gorm:"type:varchar(511);not null;default:''"`
	Text          string `json:"text,omitempty" gorm:"type:text;not null;default:''"`
}
