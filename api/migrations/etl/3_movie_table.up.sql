CREATE TABLE IF NOT EXISTS movies (
	ID bigint unsigned auto_increment,
	budget bigint unsigned null,
	tmdb_id bigint unsigned not null,
	-- original_lang_id nie ma tabeli, dodac FK pozniej.
	original_lang_id char(2) null,
	title varchar(256) not null,
	overview varchar(2048) null,
	popularity float null,
	release_date date null,
	revenue bigint null,
	runtime smallint unsigned null,
	status varchar(64) null,
	tagline varchar(128) null,
	vote_average float null,
	vote_total bigint unsigned null,
	
	PRIMARY KEY (ID),
	UNIQUE KEY (tmdb_id)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

