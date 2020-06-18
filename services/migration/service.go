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
	migration func() string
	skip      bool
}

func NewService(db *sqlx.DB, logger *logrus.Logger) Service {

	var migrations = make([]migration, 0)

	// Register Alliances Table Migration
	migrations = append(migrations, migration{
		name:      "createAllianceTable",
		migration: createAllianceTable,
		skip:      true,
	})

	// Register Characters Table Migration
	migrations = append(migrations, migration{
		name:      "createCharactersTable",
		migration: createCharactersTable,
		skip:      true,
	})

	// Register Corporations Table Migration
	migrations = append(migrations, migration{
		name:      "createCorporationsTable",
		migration: createCorporationsTable,
		skip:      true,
	})

	// Register Factions Table Migration
	migrations = append(migrations, migration{
		name:      "createFactionsTable",
		migration: createFactionsTable,
		skip:      true,
	})

	// Register Blueprint Materials Table Migration
	migrations = append(migrations, migration{
		name:      "createBlueprintMaterialsTable",
		migration: createBlueprintMaterialsTable,
		skip:      true,
	})

	// Register Blueprint Products Table Migration
	migrations = append(migrations, migration{
		name:      "createBlueprintProductsTable",
		migration: createBlueprintProductsTable,
		skip:      true,
	})

	// Register Regions Table Migration
	migrations = append(migrations, migration{
		name:      "createRegionsTable",
		migration: createRegionsTable,
		skip:      true,
	})

	// Register Constellations Table Migration
	migrations = append(migrations, migration{
		name:      "createConstellationsTable",
		migration: createConstellationsTable,
		skip:      true,
	})

	// Register Solar Systems Table Migration
	migrations = append(migrations, migration{
		name:      "createSolarSystemsTable",
		migration: createSolarSystemsTable,
		skip:      true,
	})

	// Register Types Table Migration
	migrations = append(migrations, migration{
		name:      "createTypesTable",
		migration: createTypesTable,
		skip:      true,
	})

	// Register Type Attributes Table Migration
	migrations = append(migrations, migration{
		name:      "createTypeAttributesTable",
		migration: createTypeAttributesTable,
		skip:      true,
	})

	// Register Type Categories Table Migration
	migrations = append(migrations, migration{
		name:      "createTypeCategoriesTable",
		migration: createTypeCategoriesTable,
		skip:      true,
	})

	// Register Type Flags Table Migration
	migrations = append(migrations, migration{
		name:      "createTypeFlagsTable",
		migration: createTypeFlagsTable,
		skip:      true,
	})

	// Register Type Groups Table Migration
	migrations = append(migrations, migration{
		name:      "createTypeGroupsTable",
		migration: createTypeGroupsTable,
		skip:      true,
	})

	// Register Killmails Table Migration
	migrations = append(migrations, migration{
		name:      "createKillmailsTable",
		migration: createKillmailsTable,
		skip:      true,
	})

	// Register Killmail Attackers Table Migration
	migrations = append(migrations, migration{
		name:      "createKillmailAttackersTable",
		migration: createKillmailAttackersTable,
		skip:      true,
	})

	// Register Killmail Items Table Migration
	migrations = append(migrations, migration{
		name:      "createKillmailItemsTable",
		migration: createKillmailItemsTable,
		skip:      true,
	})

	// Register Killmail Victim Table Migration
	migrations = append(migrations, migration{
		name:      "createKillmailVictimTable",
		migration: createKillmailVictimTable,
		skip:      true,
	})

	// Register Prices Table Migration
	migrations = append(migrations, migration{
		name:      "createPricesTable",
		migration: createPricesTable,
		skip:      true,
	})

	// Register Prices Built Table Migration
	migrations = append(migrations, migration{
		name:      "createPricesBuiltTable",
		migration: createPricesBuiltTable,
		skip:      true,
	})

	// Register Tokens Table Migration
	migrations = append(migrations, migration{
		name:      "createTokensTable",
		migration: createTokensTable,
		skip:      true,
	})

	// Register Alter Characters Add NoResponseCount And UpdatePriority Columns Table Migration
	migrations = append(migrations, migration{
		name:      "alterCharactersAddNoResponseCountAndUpdatePriorityColumns",
		migration: alterCharactersAddNoResponseCountAndUpdatePriorityColumns,
		skip:      true,
	})

	// Register Alter Corporations Add NoResponseCount And UpdatePriority Columns Table Migration
	migrations = append(migrations, migration{
		name:      "alterCorporationsNoResponseCountAndUpdatePriorityColumns",
		migration: alterCorporationsNoResponseCountAndUpdatePriorityColumns,
		skip:      true,
	})

	// Register Alter Alliances Add NoResponseCount And UpdatePriority Columns Table Migration
	migrations = append(migrations, migration{
		name:      "alterAlliancesNoResponseCountAndUpdatePriorityColumns",
		migration: alterAlliancesNoResponseCountAndUpdatePriorityColumns,
		skip:      true,
	})

	migrations = append(migrations, migration{
		name:      "alterCorporationsTableAddMemberCountColoumn",
		migration: alterCorporationsTableAddMemberCountColoumn,
		skip:      true,
	})

	migrations = append(migrations, migration{
		name:      "alterTableCharactersMakeEtagNullable",
		migration: alterTableCharactersMakeEtagNullable,
		skip:      true,
	})

	migrations = append(migrations, migration{
		name:      "alterTableCorporationsMakeEtagNullable",
		migration: alterTableCorporationsMakeEtagNullable,
		skip:      true,
	})

	migrations = append(migrations, migration{
		name:      "alterTableAlliancesMakeEtagNullable",
		migration: alterTableAlliancesMakeEtagNullable,
		skip:      true,
	})

	migrations = append(migrations, migration{
		name:      "updateCorporationsSetEtagNULL",
		migration: updateCorporationsSetEtagNULL,
		skip:      true,
	})

	migrations = append(migrations, migration{
		name:      "addSecStatusColumnToCharactersTable",
		migration: addSecStatusColumnToCharactersTable,
		skip:      true,
	})

	migrations = append(migrations, migration{
		name:      "dropMemberCountColumnFromAlliancesTable",
		migration: dropMemberCountColumnFromAlliancesTable,
		skip:      true,
	})
	migrations = append(migrations, migration{
		name:      "alterPricesDropPriceDefault",
		migration: alterPricesDropPriceDefault,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterPricesChangePricePercision",
		migration: alterPricesChangePricePercision,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterKillmailsDropColumnDefaults",
		migration: alterKillmailsDropColumnDefaults,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterKillmailsChangeValuePercision",
		migration: alterKillmailsChangeValuePercision,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterKillmailAttackersDropColumnDefaults",
		migration: alterKillmailAttackersDropColumnDefaults,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterKillmailAttackersChangeSecurityStatusPercision",
		migration: alterKillmailAttackersChangeSecurityStatusPercision,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterKillmailItemsDropColumnDefaults",
		migration: alterKillmailItemsDropColumnDefaults,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterKillmailItemsChangeItemValuePercision",
		migration: alterKillmailItemsChangeItemValuePercision,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterKillmailVictimDropColumnDefaults",
		migration: alterKillmailVictimDropColumnDefaults,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterKillmailVictimChangeShipValuePercision",
		migration: alterKillmailVictimChangeShipValuePercision,
		skip:      false,
	})
	migrations = append(migrations, migration{
		name:      "alterCharacterChangeSecurityStatusToDecimal",
		migration: alterCharacterChangeSecurityStatusToDecimal,
		skip:      false,
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
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if !migration.skip {
			s.logger.WithField("migration", migration.name).Info("executing migration")
			_, err = s.db.Exec(migration.migration())
			if err != nil {
				s.logger.WithError(err).WithField("migration", migration.name).Fatal("encountered error execution migration")
			}
			s.logger.WithField("migration", migration.name).Info("migration executed successfully")
		}

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
