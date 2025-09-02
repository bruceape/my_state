package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	current_time := time.Now().UTC().Format("20060102150405")
	name := "new_migration"
	if len(os.Args) > 1 {
		name = fmt.Sprintf("%s", os.Args[1])
	}

	file_name := fmt.Sprintf("db/migrations/%s_%s.sql", current_time, name)

	file, err := os.Create(file_name)
	if err != nil {
		log.Fatal("Could not create file %", file_name)
	}
	defer file.Close()

	fmt.Println("File created:", file.Name())

	file.WriteString("BEGIN;\n")
	file.WriteString("-- your SQL here \n")
	file.WriteString("COMMIT;")
}
