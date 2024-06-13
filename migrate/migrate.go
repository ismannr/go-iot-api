package main

import (
	"gin-crud/initializers"
	model "gin-crud/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.DatabaseInit()
}

func main() {
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
	}

	for _, m := range models {
		if initializers.DB.Migrator().HasTable(m.model) && initializers.DB.Migrator().HasTable(m.table) {
			initializers.DB.Migrator().DropTable(m.model)
			initializers.DB.Migrator().DropTable(m.table)
		}
	}

	for _, m := range models {
		initializers.DB.AutoMigrate(m.model)
	}
}
