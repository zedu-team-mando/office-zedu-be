package migrations

import (
	"github.com/hngprojects/telex_be/pkg/repository/storage"
	"gorm.io/gorm"
)

// FixDeletedAtColumns repairs legacy typo column names where a leading space was created
// e.g. " deleted_at", ensuring queries matching `deleted_at` succeed seamlessly.
func FixDeletedAtColumns(db *gorm.DB) {
	statements := []string{
		// 1. Rename " deleted_at" in channels if the typo column exists
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'channels' AND column_name = ' deleted_at'
			) THEN
				ALTER TABLE channels RENAME COLUMN " deleted_at" TO deleted_at;
			END IF;
		END $$;`,

		// 2. Ensure deleted_at is nullable for non-deleted records
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'channels' AND column_name = 'deleted_at' AND is_nullable = 'NO'
			) THEN
				ALTER TABLE channels ALTER COLUMN deleted_at DROP NOT NULL;
			END IF;
		END $$;`,

		// 3. Repair groups table if legacy typo column exists
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'groups' AND column_name = ' deleted_at'
			) THEN
				ALTER TABLE groups RENAME COLUMN " deleted_at" TO deleted_at;
			END IF;
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'groups' AND column_name = ' updated_at'
			) THEN
				ALTER TABLE groups RENAME COLUMN " updated_at" TO updated_at;
			END IF;
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'groups' AND column_name = 'deleted_at' AND is_nullable = 'NO'
			) THEN
				ALTER TABLE groups ALTER COLUMN deleted_at DROP NOT NULL;
			END IF;
		END $$;`,
	}

	for _, stmt := range statements {
		_ = db.Exec(stmt)
	}
}

func RunAllMigrations(db *storage.Database) {
	if db == nil || db.Postgresql == nil {
		return
	}

	// 1. First ensure legacy columns with leading space typos are repaired
	FixDeletedAtColumns(db.Postgresql)

	// 2. verification migration
	MigrateModels(db.Postgresql, AuthMigrationModels(), AlterColumnModels())
	OfficeIndexes(db.Postgresql)

}

func MigrateModels(db *gorm.DB, models []any, AlterColums []AlterColumn) {
	_ = db.AutoMigrate(models...)

	// for _, alter := range AlterColums {
	// 	if alter.Column != "" {
	// 		alter.AddColumn(db)
	// 	}
	// }
}
