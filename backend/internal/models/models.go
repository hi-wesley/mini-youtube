package models

import "time"

type User struct {
	ID        string    `gorm:"primaryKey" json:"ID"`
	Email     string    `gorm:"uniqueIndex;size:255" json:"Email"`
	Username  string    `gorm:"uniqueIndex;size:50" json:"Username"`
	AvatarURL string    `json:"AvatarURL"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type Video struct {
	ID           string    `gorm:"primaryKey" json:"ID"`
	UserID       string    `gorm:"index" json:"UserID"`
	Title        string    `gorm:"size:120" json:"Title"`
	Description  string    `gorm:"type:text" json:"Description"`
	ObjectName   string    `json:"ObjectName"`
	Summary      string    `gorm:"type:text" json:"Summary"`
	SummaryModel string    `gorm:"size:50" json:"SummaryModel"`
	Views        int64     `json:"Views"`
	CreatedAt    time.Time `json:"CreatedAt"`
	User         User      `gorm:"foreignKey:UserID" json:"User"`
	Comments     []Comment `json:"Comments"`
	Likes        int       `gorm:"-" json:"Likes"`
	IsLiked      bool      `gorm:"-" json:"IsLiked"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"ID"`
	UserID    string    `json:"UserID"`
	VideoID   string    `gorm:"index" json:"VideoID"`
	Message   string    `gorm:"type:text" json:"Message"`
	CreatedAt time.Time `json:"CreatedAt"`
	User      User      `gorm:"foreignKey:UserID" json:"User"`
}

type Like struct {
	UserID  string `gorm:"primaryKey" json:"UserID"`
	VideoID string `gorm:"primaryKey" json:"VideoID"`
}
