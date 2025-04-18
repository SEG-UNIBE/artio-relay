package models

/*
Log Type definition of the type object in the database
*/
type Log struct {
	AModel
	IP      string
	Type    string `gorm:"not null"`
	Content string `gorm:"not null"`
}
