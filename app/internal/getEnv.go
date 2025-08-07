package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load("/app/.env")
	// err := godotenv.Load("../docker/.env")
	if err != nil {
		fmt.Println("Ошибка:", err)
		relativePath := "."
		absolutePath, err := filepath.Abs(relativePath)
		if err != nil {
			log.Println("Ошибка:", err)
		}
		log.Println("Абсолютный путь:", absolutePath)
		files, _ := os.ReadDir(absolutePath)
		fmt.Printf("Файлы в %s:\n", absolutePath)
		for _, file := range files {
			fmt.Println(file.Name())
		}
		log.Fatal("error loading .env file")
		return fmt.Errorf("error loading .env file: %v", err)
	}
	return nil
}
