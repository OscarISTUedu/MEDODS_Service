package models

import (
	"sync"

	"gorm.io/gorm"
)

var (
	registeredModels []interface{}
	modelMu          sync.Mutex
)

func registerModel(model interface{}) {
	modelMu.Lock()
	defer modelMu.Unlock()
	registeredModels = append(registeredModels, model)
}

func GetRegisteredModels() []interface{} {
	return registeredModels
}

type User struct {
	gorm.Model
	Login    string `gorm:"size:255;not null" json:"login"`
	Password string `gorm:"size:255;not null" json:"-"`
}

type RefreshToken struct {
	gorm.Model
	Token     string `gorm:"uniqueIndex;size:255"`
	UserID    uint
	User      User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserAgent string
	IpAdress  string
}

func init() {
	registerModel(&User{})
	registerModel(&RefreshToken{})
}
