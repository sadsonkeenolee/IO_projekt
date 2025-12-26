create table if not exists genres (
	ID bigint unsigned auto_increment,
	genre varchar(32) not null,
	primary key (ID),
	unique key (genre)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists keywords (
	ID bigint unsigned auto_increment,
	keyword varchar(32) not null,
	primary key (ID),
	unique key (keyword)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists languages (
	encoding char(2) not null,
	language varchar(32) not null,
	primary key (encoding),
	unique key (language)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists companies (
	ID bigint unsigned auto_increment,
	company varchar(32) not null,
	primary key (ID),
	unique key (company)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists countries (
	encoding char(2),
	country varchar(32) not null,
	primary key (encoding),
	unique key (country)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;
