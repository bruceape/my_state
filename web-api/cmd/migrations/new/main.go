package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	currentTime := time.Now().UTC().Format("20060102150405")
	name := "new_migration"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	dir := "db/migrations"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("could not create migrations dir %s: %v", dir, err)
	}

	fileName := fmt.Sprintf("%s/%s_%s.sql", dir, currentTime, name)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Could not create file %s: %v", fileName, err)
	}
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		log.Fatalf("Failed to close file %v", err)
	}

	fmt.Println("File created:", file.Name())

	if _, err = file.WriteString("-- your SQL here \n"); err != nil {
		log.Fatalf("write %s: %v", fileName, err)
	}
}
