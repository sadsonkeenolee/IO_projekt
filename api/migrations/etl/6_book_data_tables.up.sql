CREATE TABLE IF NOT EXISTS books (
	ID bigint unsigned auto_increment,
	title varchar(128) not null,
	rating float not null,
	isbn char(10) not null,
	isbn13 char(13) not null,
	language_code varchar(3),
	pages int,
	total_ratings int,
	release_date date null,
	publisher varchar(128),
	PRIMARY KEY (ID)
	-- UNIQUE KEY (movie_id, genre_id),
	-- FOREIGN KEY (`movie_id`) REFERENCES movies(`tmdb_id`) ON DELETE CASCADE,
	-- FOREIGN KEY (`genre_id`) REFERENCES genres(`ID`) ON DELETE CASCADE

)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS authors (
	ID bigint unsigned auto_increment,
	book_id bigint unsigned,
	author varchar(128) not null,
	PRIMARY KEY (ID),
	FOREIGN KEY (`book_id`) REFERENCES books(`ID`) ON DELETE CASCADE

)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;
