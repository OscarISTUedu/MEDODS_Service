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
	Login    string `gorm:"size:255;not null" json:"username"`
	Password string `gorm:"size:255;not null" json:"-"`
}

func init() {
	registerModel(&User{})
}
