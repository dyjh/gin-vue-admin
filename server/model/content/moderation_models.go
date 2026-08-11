package content

// ModerationPersistenceModels is the shared AutoMigrate integration point.
func ModerationPersistenceModels() []interface{} {
	return []interface{}{
		&ModerationConfig{},
		&ModerationConnectionTest{},
		&ModerationImageTest{},
		&ModerationConfigHealth{},
		&ImageModerationRecord{},
	}
}
