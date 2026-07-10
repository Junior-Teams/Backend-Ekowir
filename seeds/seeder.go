package seeds

import (
	"flag"
	"fmt"
	// "strconv"

	// "github.com/ALZEE23/ApiGo/database"
	// "github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/seeds/tables"
)

func RunSeeders() {
	seed := flag.String("seed", "", "Run specific seeder")
	flag.Parse()

	switch *seed {
	case "users":
		tables.SeedUsers()
	default:
		fmt.Println("No valid seeder specified")
	}
}