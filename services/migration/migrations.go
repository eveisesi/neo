package migration

func createAllianceTable() string {
	return `
		CREATE TABLE IF NOT EXISTS alliances (
			id bigint(20) unsigned NOT NULL,
			name varchar(255) NOT NULL,
			ticker varchar(5) NOT NULL,
			member_count bigint(20) unsigned NOT NULL DEFAULT '0',
			is_closed tinyint(1) NOT NULL DEFAULT '0',
			etag varchar(255) NOT NULL,
			cached_until datetime NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id),
			KEY alliances_cached_until (cached_until)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`
}

func createBlueprintMaterialsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS blueprint_materials (
			type_id bigint(20) unsigned NOT NULL,
			activity_id bigint(20) unsigned NOT NULL,
			material_type_id bigint(20) unsigned NOT NULL,
			quantity bigint(20) unsigned NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (type_id,activity_id,material_type_id),
			KEY blueprint_materials_type_id (type_id),
			KEY blueprint_materials_material_type_id (material_type_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`
}

func createBlueprintProductsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS blueprint_products (
			type_id bigint(20) unsigned NOT NULL,
			activity_id bigint(20) unsigned NOT NULL,
			product_type_id bigint(20) unsigned NOT NULL,
			quantity bigint(20) unsigned NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (type_id,activity_id,product_type_id),
			KEY blueprint_products_type_id (type_id),
			KEY blueprint_products_product_type_id (product_type_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`
}

func createCharactersTable() string {
	return `
		CREATE TABLE IF NOT EXISTS characters (
			id bigint(20) unsigned NOT NULL,
			name varchar(255) NOT NULL,
			corporation_id bigint(20) unsigned NOT NULL,
			alliance_id bigint(20) unsigned DEFAULT NULL,
			faction_id bigint(20) unsigned DEFAULT NULL,
			etag varchar(255) NOT NULL,
			cached_until datetime NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id)
			INDEX characters_cached_until (cached_until)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`
}

func createConstellationsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS constellations (
			id bigint(20) unsigned NOT NULL,
			name varchar(100) NOT NULL,
			region_id bigint(20) unsigned NOT NULL,
			pos_x double NOT NULL,
			pos_y double NOT NULL,
			pos_z double NOT NULL,
			faction_id bigint(20) unsigned DEFAULT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`
}

func createCorporationsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS corporations (
			id bigint(20) unsigned NOT NULL,
			name varchar(255) NOT NULL,
			ticker varchar(10) NOT NULL,
			alliance_id bigint(20) unsigned DEFAULT NULL,
			etag varchar(255) NOT NULL,
			cached_until datetime NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id),
			KEY corporations_cached_until (cached_until)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createFactionsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS factions (
			id bigint(20) unsigned NOT NULL,
			name varchar(100) NOT NULL,
			description varchar(1000) NOT NULL,
			race_id bigint(20) unsigned NOT NULL,
			solar_system_id bigint(20) unsigned NOT NULL,
			corporation_id bigint(20) unsigned DEFAULT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createKillmailsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS killmails (
			id bigint(20) unsigned NOT NULL,
			hash varchar(255) NOT NULL,
			moon_id bigint(20) DEFAULT NULL,
			solar_system_id bigint(20) unsigned NOT NULL,
			war_id bigint(20) DEFAULT NULL,
			is_npc tinyint(1) NOT NULL DEFAULT '0',
			is_awox tinyint(1) NOT NULL DEFAULT '0',
			is_solo tinyint(1) NOT NULL DEFAULT '0',
			dropped_value decimal(18,4) unsigned NOT NULL DEFAULT '0.0000',
			destroyed_value decimal(18,4) unsigned NOT NULL DEFAULT '0.0000',
			fitted_value decimal(18,4) unsigned NOT NULL DEFAULT '0.0000',
			total_value decimal(18,4) unsigned NOT NULL DEFAULT '0.0000',
			killmail_time datetime NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id,hash),
			KEY total_value (total_value),
			KEY killmail_time (killmail_time),
			KEY solar_system_id (solar_system_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`
}

func createKillmailAttackersTable() string {

	return `
		CREATE TABLE IF NOT EXISTS killmail_attackers (
			id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
			killmail_id bigint(20) unsigned NOT NULL,
			alliance_id bigint(20) unsigned DEFAULT NULL,
			character_id bigint(20) unsigned DEFAULT NULL,
			corporation_id bigint(20) unsigned DEFAULT NULL,
			faction_id bigint(20) unsigned DEFAULT NULL,
			damage_done bigint(20) unsigned NOT NULL,
			final_blow tinyint(1) NOT NULL DEFAULT '0',
			security_status decimal(17,15) NOT NULL,
			ship_type_id bigint(20) unsigned DEFAULT NULL,
			weapon_type_id bigint(20) unsigned DEFAULT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id),
			KEY ship_type_id (ship_type_id),
			KEY weapon_type_id (weapon_type_id),
			KEY alliance_id (alliance_id),
			KEY corporation_id (corporation_id),
			KEY killmail_id_character_id (killmail_id,character_id),
			CONSTRAINT killmail_attackers_killmail_id_killmails_id_foreign_key FOREIGN KEY (killmail_id) REFERENCES killmails (id) ON DELETE CASCADE ON UPDATE CASCADE
		) ENGINE=InnoDB AUTO_INCREMENT=5552301 DEFAULT CHARSET=utf8;
	  `
}

func createKillmailItemsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS killmail_items (
			id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
			parent_id bigint(20) unsigned DEFAULT NULL,
			killmail_id bigint(20) unsigned NOT NULL,
			flag bigint(20) unsigned NOT NULL,
			item_type_id bigint(20) unsigned NOT NULL,
			quantity_dropped bigint(20) unsigned DEFAULT NULL,
			quantity_destroyed bigint(20) unsigned DEFAULT NULL,
			item_value decimal(18,4) NOT NULL DEFAULT '0.0000',
			singleton bigint(20) unsigned NOT NULL,
			is_parent tinyint(1) NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id),
			KEY killmail_id (killmail_id),
			KEY item_type_id (item_type_id),
			KEY parent_id (parent_id),
			KEY flag_id (flag),
			CONSTRAINT killmail_items_killmail_id_killmails_id_foreign_key FOREIGN KEY (killmail_id) REFERENCES killmails (id) ON DELETE CASCADE ON UPDATE CASCADE
		) ENGINE=InnoDB AUTO_INCREMENT=17910512 DEFAULT CHARSET=utf8;
	`
}

func createKillmailVictimTable() string {
	return `
		CREATE TABLE IF NOT EXISTS killmail_victim (
			id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
			killmail_id bigint(20) unsigned NOT NULL,
			alliance_id bigint(20) unsigned DEFAULT NULL,
			character_id bigint(20) unsigned DEFAULT NULL,
			corporation_id bigint(20) unsigned DEFAULT NULL,
			faction_id bigint(20) unsigned DEFAULT NULL,
			damage_taken bigint(20) unsigned NOT NULL,
			ship_type_id bigint(20) unsigned NOT NULL,
			ship_value decimal(18,4) NOT NULL DEFAULT '0.0000',
			pos_x decimal(30,10) DEFAULT NULL,
			pos_y decimal(30,10) DEFAULT NULL,
			pos_z decimal(30,10) DEFAULT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id),
			KEY corporation_id (corporation_id),
			KEY alliance_id (alliance_id),
			KEY ship_type_id (ship_type_id),
			KEY killmail_id_character_id (killmail_id,character_id),
			CONSTRAINT killmail_victim_killmail_id_killmails_id_foreign_key FOREIGN KEY (killmail_id) REFERENCES killmails (id) ON DELETE CASCADE ON UPDATE CASCADE
		) ENGINE=InnoDB AUTO_INCREMENT=1337219 DEFAULT CHARSET=utf8;
	  `
}

func createPricesTable() string {
	return `
		CREATE TABLE IF NOT EXISTS prices (
			type_id bigint(20) unsigned NOT NULL,
			date date NOT NULL,
			price decimal(18,4) unsigned NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (type_id,date),
			KEY date (date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createPricesBuiltTable() string {
	return `
		CREATE TABLE IF NOT EXISTS prices_built (
			type_id bigint(20) unsigned NOT NULL,
			date date NOT NULL,
			price decimal(18,4) NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (type_id,date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createRegionsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS regions (
			id bigint(20) unsigned NOT NULL,
			name varchar(100) NOT NULL,
			pos_x double NOT NULL,
			pos_y double NOT NULL,
			pos_z double NOT NULL,
			faction_id int(11) unsigned DEFAULT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createSolarSystemsTable() string {
	return `	
		CREATE TABLE IF NOT EXISTS solar_systems (
			id bigint(20) unsigned NOT NULL,
			name varchar(100) NOT NULL,
			constellation_id bigint(20) unsigned NOT NULL,
			faction_id bigint(20) unsigned DEFAULT NULL,
			sun_type_id bigint(20) unsigned DEFAULT NULL,
			pos_x double NOT NULL,
			pos_y double NOT NULL,
			pos_z double NOT NULL,
			security double NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id),
			KEY ix_mapSolarSystems_constellationID (constellation_id),
			KEY ix_mapSolarSystems_security (security)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createTokensTable() string {
	return `
		CREATE TABLE IF NOT EXISTS tokens (
			id bigint(20) unsigned NOT NULL,
			main bigint(20) unsigned NOT NULL,
			access_token text NOT NULL,
			refresh_token varchar(255) NOT NULL,
			expiry datetime NOT NULL,
			disabled tinyint(1) NOT NULL DEFAULT '0',
			disabled_timestamp datetime DEFAULT NULL,
			disabled_reason varchar(255) NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id),
			KEY disabled (disabled)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createTypesTable() string {
	return `	
		CREATE TABLE IF NOT EXISTS types (
			id bigint(20) unsigned NOT NULL,
			group_id bigint(20) unsigned NOT NULL,
			name varchar(100) NOT NULL,
			description text NOT NULL,
			published tinyint(1) NOT NULL,
			market_group_id bigint(20) unsigned DEFAULT NULL,
			created_at datetime DEFAULT CURRENT_TIMESTAMP,
			updated_at datetime DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY ix_invTypes_groupID (group_id),
			KEY market_group_id (market_group_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createTypeAttributesTable() string {
	return `	
		CREATE TABLE IF NOT EXISTS type_attributes (
			type_id bigint(20) unsigned NOT NULL,
			attribute_id bigint(20) unsigned NOT NULL,
			value bigint(20) NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (type_id,attribute_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createTypeCategoriesTable() string {
	return `	
		CREATE TABLE IF NOT EXISTS type_categories (
			id bigint(20) unsigned NOT NULL,
			name varchar(255) NOT NULL,
			published tinyint(1) NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createTypeFlagsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS type_flags (
			id bigint(20) unsigned NOT NULL,
			name varchar(200) NOT NULL,
			text varchar(100) NOT NULL,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func createTypeGroupsTable() string {
	return `
		CREATE TABLE IF NOT EXISTS type_groups (
			id bigint(20) unsigned NOT NULL,
			category_id bigint(20) unsigned NOT NULL,
			name varchar(255) NOT NULL,
			published tinyint(1) NOT NULL DEFAULT '0',
			created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY ix_invGroups_categoryID (category_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	  `
}

func alterCharactersAddNoResponseCountAndUpdatePriorityColumns() string {
	return `
		ALTER TABLE characters
			ADD COLUMN not_modified_count INT UNSIGNED NOT NULL DEFAULT '0' AFTER faction_id,
			ADD COLUMN update_priority INT UNSIGNED NOT NULL DEFAULT '0' AFTER not_modified_count;
	`
}

func alterCorporationsNoResponseCountAndUpdatePriorityColumns() string {
	return `
		ALTER TABLE corporations
			ADD COLUMN not_modified_count INT UNSIGNED NOT NULL DEFAULT '0' AFTER alliance_id,
			ADD COLUMN update_priority INT UNSIGNED NOT NULL DEFAULT '0' AFTER not_modified_count;
	`
}

func alterAlliancesNoResponseCountAndUpdatePriorityColumns() string {
	return `
		ALTER TABLE alliances
			ADD COLUMN not_modified_count INT UNSIGNED NOT NULL DEFAULT '0' AFTER is_closed,
			ADD COLUMN update_priority INT UNSIGNED NOT NULL DEFAULT '0' AFTER not_modified_count;
	`
}

func alterCorporationsTableAddMemberCountColoumn() string {
	return `
		ALTER TABLE corporations
			ADD COLUMN member_count INT UNSIGNED NOT NULL DEFAULT 0 AFTER ticker;
	`
}

func alterTableCharactersMakeEtagNullable() string {
	return `
		ALTER TABLE characters
			CHANGE COLUMN etag etag VARCHAR(255) NULL AFTER update_priority;
	`
}

func alterTableCorporationsMakeEtagNullable() string {
	return `
		ALTER TABLE corporations
			CHANGE COLUMN etag etag VARCHAR(255) NULL AFTER update_priority;
	`
}

func alterTableAlliancesMakeEtagNullable() string {
	return `
		ALTER TABLE alliances
			CHANGE COLUMN etag etag VARCHAR(255) NULL AFTER update_priority;
	`
}

func updateCorporationsSetEtagNULL() string {
	return `
		UPDATE corporations SET etag = NULL
	`
}

func addSecStatusColumnToCharactersTable() string {
	return `
		ALTER TABLE characters
			ADD COLUMN security_status DOUBLE NOT NULL DEFAULT '0.00' AFTER name;
	`
}

func dropMemberCountColumnFromAlliancesTable() string {
	return `
		ALTER TABLE alliances
			DROP COLUMN member_count;
	`
}

func alterPricesDropPriceDefault() string {
	return `
		ALTER TABLE prices
			ALTER price DROP DEFAULT;
	`
}

func alterPricesChangePricePercision() string {
	return `
		ALTER TABLE prices
			CHANGE COLUMN price price DECIMAL(18,2) UNSIGNED NOT NULL DEFAULT 0.00 AFTER date;
			
	`
}

func alterKillmailAttackersChangeSecurityStatusPercision() string {
	return `
		ALTER TABLE killmail_attackers
			CHANGE COLUMN security_status security_status DECIMAL(4,2) NOT NULL AFTER final_blow;
	`
}

func alterCharacterChangeSecurityStatusToDecimal() string {
	return `
		ALTER TABLE characters
			CHANGE COLUMN security_status security_status DECIMAL(4,2) NOT NULL DEFAULT '0' AFTER name;
	`
}
