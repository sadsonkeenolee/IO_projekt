create table if not exists movies (
	ID bigint unsigned auto_increment,
	budget bigint unsigned null,
	tmdb_id bigint unsigned not null,
	language char(2) null,
	title varchar(256) not null,
	overview varchar(2048) null,
	popularity float null,
	release_date date null,
	revenue bigint null,
	runtime smallint unsigned null,
	status enum ('Released', 'Rumored', 'Post Production', 'N/A') default 'N/A',
	tagline varchar(128) null,
	rating float default 0,
	total_ratings bigint unsigned default 0,
	
	primary key (ID),
	unique key (tmdb_id)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

