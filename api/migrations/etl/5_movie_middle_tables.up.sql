CREATE TABLE IF NOT EXISTS movie2genres (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned,
	genre_id bigint unsigned,
	
	PRIMARY KEY (ID),
	UNIQUE KEY (movie_id, genre_id),
	FOREIGN KEY (`movie_id`) REFERENCES movies(`tmdb_id`) ON DELETE CASCADE,
	FOREIGN KEY (`genre_id`) REFERENCES genres(`ID`) ON DELETE CASCADE

)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS movie2keywords (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned,
	keyword_id bigint unsigned,
	
	PRIMARY KEY (ID),
	UNIQUE KEY (movie_id, keyword_id),
	FOREIGN KEY (`movie_id`) REFERENCES movies(`tmdb_id`) ON DELETE CASCADE,
	FOREIGN KEY (`keyword_id`) REFERENCES keywords(`ID`) ON DELETE CASCADE

)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS movie2languages (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned,
	language_id char(2),
	
	PRIMARY KEY (ID),
	UNIQUE KEY (movie_id, language_id),
	FOREIGN KEY (`movie_id`) REFERENCES movies(`tmdb_id`) ON DELETE CASCADE,
	FOREIGN KEY (`language_id`) REFERENCES languages(`encoding`) ON DELETE CASCADE

)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS movie2companies (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned,
	company_id bigint unsigned,
	
	PRIMARY KEY (ID),
	UNIQUE KEY (movie_id, company_id),
	FOREIGN KEY (`movie_id`) REFERENCES movies(`tmdb_id`) ON DELETE CASCADE,
	FOREIGN KEY (`company_id`) REFERENCES companies(`ID`) ON DELETE CASCADE

)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS movie2countries (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned not null,
	country_en char(2) not null,
	
	PRIMARY KEY (ID),
	UNIQUE KEY (movie_id, country_en),
	FOREIGN KEY (`movie_id`) REFERENCES movies(`tmdb_id`) ON DELETE CASCADE,
	FOREIGN KEY (`country_en`) REFERENCES countries(`encoding`) ON DELETE CASCADE

)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE movies
	ADD CONSTRAINT FOREIGN KEY(original_lang_id) REFERENCES languages(encoding);
