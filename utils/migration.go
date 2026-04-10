package utils

import (
	"gorm.io/gorm"
)

func AddPermissionCompositeIndex(db *gorm.DB) error {
	var indexExists bool
	err := db.Raw(`
		SELECT COUNT(*) > 0 
		FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = 'inpatient_permissions' 
		AND index_name = 'idx_role_module_action'
	`).Scan(&indexExists).Error

	if err != nil {
		return err
	}
	if !indexExists {
		return db.Exec(`
			CREATE INDEX idx_role_module_action 
			ON inpatient_permissions(role_id, module_id, action)
		`).Error
	}

	return nil
}
