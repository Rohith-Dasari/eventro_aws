package models

type Artist struct {
	ID   string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name string `gorm:"type:text;not null" json:"name"`
	Bio  string `gorm:"type:text" json:"bio"`
}
