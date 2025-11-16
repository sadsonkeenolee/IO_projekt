CREATE TABLE IF NOT EXISTS genres (
	ID bigint unsigned auto_increment,
	name varchar(32) not null,
	PRIMARY KEY (ID)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS keywords (
	ID bigint unsigned auto_increment,
	name varchar(32) not null,
	
	PRIMARY KEY (ID)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS languages (
	ID bigint unsigned auto_increment,
	name varchar(32) not null,
	
	PRIMARY KEY (ID)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS companies (
	ID bigint unsigned auto_increment,
	name varchar(32) not null,
	
	PRIMARY KEY (ID)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS countries (
	ID bigint unsigned auto_increment,
	name varchar(32) not null,
	
	PRIMARY KEY (ID)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;
