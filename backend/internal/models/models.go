package models

import "time"

type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;size:255"`
	Username  string `gorm:"uniqueIndex;size:50"`
	AvatarURL string
	CreatedAt time.Time
}

type Video struct {
	ID          string `gorm:"primaryKey"`
	UserID      string `gorm:"index"`
	Title       string `gorm:"size:120"`
	Description string `gorm:"type:text"`
	ObjectName  string // in GCS: "videos/uid/filename.mp4"
	Summary     string `gorm:"type:text"`
	Views       int64
	CreatedAt   time.Time
	User        User `gorm:"foreignKey:UserID"`
	Comments    []Comment
	Likes       []Like
}

type Comment struct {
	ID        uint `gorm:"primaryKey"`
	UserID    string
	VideoID   string `gorm:"index"`
	Message   string `gorm:"type:text"`
	CreatedAt time.Time
	User      User `gorm:"foreignKey:UserID"`
}

type Like struct {
	UserID  string `gorm:"primaryKey"`
	VideoID string `gorm:"primaryKey"`
}
