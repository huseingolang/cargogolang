package Inerfaces

import "gorm.io/gorm"

type IHandler interface {
	Init() *gorm.DB
}
type Connect struct {
	Db *gorm.DB
}
