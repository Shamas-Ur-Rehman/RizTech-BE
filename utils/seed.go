package utils

import (
	"log"

	"supergit/inpatient/models"

	"gorm.io/gorm"
)

func SeedRBAC(db *gorm.DB) error {
	modules := []models.Module{
		{Name: "user", DisplayName: "User Management", Description: "Manage users in the system"},
		{Name: "patient", DisplayName: "Patient Management", Description: "Manage patient information"},
		{Name: "role", DisplayName: "Role Management", Description: "Manage roles and permissions"},
		{Name: "book", DisplayName: "Book Management", Description: "Manage books (multi-tenancy test)"},
	}

	for i := range modules {
		var m models.Module
		if err := db.Where("name = ?", modules[i].Name).FirstOrCreate(&m, modules[i]).Error; err != nil {
			return err
		}
		modules[i].ID = m.ID
	}
	roles := []models.Role{
		{Name: "admin"},
	}

	for i := range roles {
		var r models.Role
		if err := db.Where("name = ?", roles[i].Name).FirstOrCreate(&r, roles[i]).Error; err != nil {
			return err
		}
		roles[i].ID = r.ID
	}
	actions := []string{"get", "create", "update", "delete"}

	for _, module := range modules {
		for _, action := range actions {
			var p models.Permission
			if err := db.Where("role_id = ? AND module_id = ? AND action = ?", roles[0].ID, module.ID, action).
				FirstOrCreate(&p, models.Permission{
					RoleID:   roles[0].ID,
					ModuleID: module.ID,
					Action:   action,
				}).Error; err != nil {
				return err
			}
		}
	}
	log.Println("seed completed with modules, roles, permissions, and admin user")
	return nil
}
