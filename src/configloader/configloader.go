package configloader

import (
	"log"
	// "os" // Uncomment if you need to construct absolute paths for .env
	// "path/filepath" // Uncomment if you need to construct absolute paths for .env

	"github.com/joho/godotenv"
)

func init() {
	// Attempt to load .env file.
	// This assumes CourseTool.env is in the same directory as the executable,
	// or in the working directory from which the executable is run.
	// When you run E:\DEV\Go\CourseTool\temp\CourseTool.exe,
	// and CourseTool.env is also in E:\DEV\Go\CourseTool\temp\, this will find it.
	err := godotenv.Load("CourseTool.env")
	if err != nil {
		// It's common for .env files to be optional, especially in production
		// where env vars are set directly. So, a warning is often sufficient.
		log.Printf("CONFIGLOADER: Note: Error loading CourseTool.env file: %v. Will rely on system-set environment variables if they are present.", err)
	} else {
		log.Println("CONFIGLOADER: CourseTool.env loaded successfully.")
	}
}
