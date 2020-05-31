package migration

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Init() error
	Run()
}

type service struct {
	db         *sqlx.DB
	logger     *logrus.Logger
	migrations []migration
}

type migration struct {
	name      string
	migration func(db *sqlx.DB) error
}

func NewService(db *sqlx.DB, logger *logrus.Logger) Service {

	var migrations = make([]migration, 0)

	// Register Alliances Table Migration
	migrations = append(migrations, migration{
		name:      "create_alliances_table",
		migration: createAllianceTable,
	})

	// Register Characters Table Migration
	migrations = append(migrations, migration{
		name:      "create_characters_table",
		migration: createCharactersTable,
	})

	// Register Corporations Table Migration
	migrations = append(migrations, migration{
		name:      "create_corporations_table",
		migration: createCorporationsTable,
	})

	// Register Factions Table Migration
	migrations = append(migrations, migration{
		name:      "create_factions_table",
		migration: createFactionsTable,
	})

	// Register Blueprint Materials Table Migration
	migrations = append(migrations, migration{
		name:      "create_blueprint_materials_table",
		migration: createBlueprintMaterialsTable,
	})

	// Register Blueprint Products Table Migration
	migrations = append(migrations, migration{
		name:      "create_blueprint_products_table",
		migration: createBlueprintProductsTable,
	})

	// Register Regions Table Migration
	migrations = append(migrations, migration{
		name:      "create_regions_table",
		migration: createRegionsTable,
	})

	// Register Constellations Table Migration
	migrations = append(migrations, migration{
		name:      "create_constellations_table",
		migration: createConstellationsTable,
	})

	// Register Solar Systems Table Migration
	migrations = append(migrations, migration{
		name:      "create_solar_systems_table",
		migration: createSolarSystemsTable,
	})

	// Register Types Table Migration
	migrations = append(migrations, migration{
		name:      "create_types_table",
		migration: createTypesTable,
	})

	// Register Type Attributes Table Migration
	migrations = append(migrations, migration{
		name:      "create_type_attributes_table",
		migration: createTypeAttributesTable,
	})

	// Register Type Categories Table Migration
	migrations = append(migrations, migration{
		name:      "create_type_categories_table",
		migration: createTypeCategoriesTable,
	})

	// Register Type Flags Table Migration
	migrations = append(migrations, migration{
		name:      "create_type_flags_table",
		migration: createTypeFlagsTable,
	})

	// Register Type Groups Table Migration
	migrations = append(migrations, migration{
		name:      "create_type_groups_table",
		migration: createTypeGroupsTable,
	})

	// Register Killmails Table Migration
	migrations = append(migrations, migration{
		name:      "create_killmails_table",
		migration: createKillmailsTable,
	})

	// Register Killmail Attackers Table Migration
	migrations = append(migrations, migration{
		name:      "create_killmail_attackers_table",
		migration: createKillmailAttackersTable,
	})

	// Register Killmail Items Table Migration
	migrations = append(migrations, migration{
		name:      "create_killmail_items_table",
		migration: createKillmailItemsTable,
	})

	// Register Killmail Victim Table Migration
	migrations = append(migrations, migration{
		name:      "create_killmail_victim_table",
		migration: createKillmailVictimTable,
	})

	// Register Prices Table Migration
	migrations = append(migrations, migration{
		name:      "create_prices_table",
		migration: createPricesTable,
	})

	// Register Prices Built Table Migration
	migrations = append(migrations, migration{
		name:      "create_prices_built_table",
		migration: createPricesBuiltTable,
	})

	// Register Tokens Table Migration
	migrations = append(migrations, migration{
		name:      "create_tokens_table",
		migration: createTokensTable,
	})

	// Register Alter Characters Add NoResponseCount And UpdatePriority Columns Table Migration
	migrations = append(migrations, migration{
		name:      "alter_characters_add_no_response_count_and_update_priority_columns",
		migration: alterCharactersAddNoResponseCountAndUpdatePriorityColumns,
	})

	// Register Alter Corporations Add NoResponseCount And UpdatePriority Columns Table Migration
	migrations = append(migrations, migration{
		name:      "alter_ccorporations_add_no_response_count_and_update_priority_columns",
		migration: alterCorporationsNoResponseCountAndUpdatePriorityColumns,
	})

	// Register Alter Alliances Add NoResponseCount And UpdatePriority Columns Table Migration
	migrations = append(migrations, migration{
		name:      "alter_alliances_add_no_response_count_and_update_priority_columns",
		migration: alterAlliancesNoResponseCountAndUpdatePriorityColumns,
	})

	migrations = append(migrations, migration{
		name:      "alter_corporations_table_add_member_count_column",
		migration: alterCorporationsTableAddMemberCountColoumn,
	})

	migrations = append(migrations, migration{
		name:      "alterTablesMakeEtagNullable",
		migration: alterTablesMakeEtagNullable,
	})

	migrations = append(migrations, migration{
		name:      "updateCorporationsSetEtagNULL",
		migration: updateCorporationsSetEtagNULL,
	})

	migrations = append(migrations, migration{
		name:      "addSecStatusColumnToCharactersTable",
		migration: addSecStatusColumnToCharactersTable,
	})

	return &service{
		db:         db,
		logger:     logger,
		migrations: migrations,
	}

}

func (s *service) Init() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
			migration VARCHAR(255) NOT NULL,
			created_at DATETIME NOT NULL,
			PRIMARY KEY (id),
			UNIQUE INDEX migration (migration)
		)
		ENGINE=InnoDB
		;
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *service) Run() {

	for _, migration := range s.migrations {

		run, err := s.hasMigrationRun(migration.name)
		if err != nil {
			s.logger.WithError(err).WithField("migration", migration.name).Error("unable to determine if migration has run")
			return
		}
		if run {

			time.Sleep(time.Millisecond * 250)
			continue
		}

		s.logger.WithField("migration", migration.name).Info("executing migration")
		err = migration.migration(s.db)
		if err != nil {
			s.logger.WithError(err).WithField("migration", migration.name).Fatal("encountered error execution migration")
		}
		s.logger.WithField("migration", migration.name).Info("migration executed successfully")

		s.logger.WithField("migration", migration.name).Info("registering migration")
		err = s.registerMigration(migration.name)
		if err != nil {
			s.logger.WithError(err).WithField("migration", migration.name).Fatal("failed to register migration")
		}
		s.logger.WithField("migration", migration.name).Info("migration registered successfully")
		time.Sleep(time.Millisecond * 250)
	}

}

func (s *service) hasMigrationRun(name string) (bool, error) {

	query := `
		SELECT COUNT(migration) from migrations where migration = ?
	`
	var count int
	err := s.db.Get(&count, query, name)
	if err != nil {
		return false, err
	}

	return count > 0, nil

}

func (s *service) registerMigration(name string) error {

	query := `
		INSERT INTO migrations (
			migration,
			created_at
		)  VALUE (
			?,
			NOW()
		)
	`

	_, err := s.db.Exec(query, name)

	return err

}
