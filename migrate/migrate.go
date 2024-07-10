package main

import (
	"gin-crud/initializers"
	model "gin-crud/models"
	"log"
)

func main() {
	initializers.LoadEnvVariables()
	initializers.DatabaseInit()

	models := []struct {
		model interface{}
		table string
	}{
		{&model.UmkmData{}, "umkm_data"},
		{&model.SystemData{}, "system_data"},
		{&model.BinusianData{}, "binusian_data"},
		{&model.Token{}, "tokens"},
		{&model.PasswordRecoveryToken{}, "password_recovery_tokens"},
		{&model.Device{}, "devices"},
		{&model.DeviceGrouping{}, "device_grouping"},
	}

	for _, m := range models {
		if initializers.DB.Migrator().HasTable(m.model) {
			if err := initializers.DB.Migrator().DropTable(m.model); err != nil {
				log.Printf("Error dropping table %s: %v\n", m.table, err)
			}
		}
	}

	for _, m := range models {
		if err := initializers.DB.AutoMigrate(m.model); err != nil {
			log.Printf("Error migrating model %T: %v\n", m.model, err)
		}
	}
	log.Println("Database migration completed successfully")
}
