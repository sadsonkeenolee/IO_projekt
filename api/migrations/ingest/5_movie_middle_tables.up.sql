create table if not exists movie2genres (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned,
	genre_id bigint unsigned,
	
	primary key (ID),
	unique key (movie_id, genre_id),
	foreign key (movie_id) references movies(tmdb_id) on delete cascade,
	foreign key (genre_id) references genres(ID) on delete cascade
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists movie2keywords (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned,
	keyword_id bigint unsigned,
	
	primary key (ID),
	unique key (movie_id, keyword_id),
	foreign key (movie_id) references movies(tmdb_id) on delete cascade,
	foreign key (keyword_id) references keywords(ID) on delete cascade
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists movie2languages (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned,
	language_encoding char(2),
	
	primary key (ID),
	unique key (movie_id, language_encoding),
	foreign key (`movie_id`) references movies(`tmdb_id`) on delete cascade,
	foreign key (`language_encoding`) REFERENCES languages(`encoding`) on delete cascade
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists movie2companies (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned,
	company_id bigint unsigned,
	
	primary key (ID),
	unique key (movie_id, company_id),
	foreign key (`movie_id`) references movies(`tmdb_id`) on delete cascade,
	foreign key (`company_id`) references companies(`ID`) on delete cascade
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists movie2countries (
	ID bigint unsigned auto_increment,
	movie_id bigint unsigned not null,
	country_encoding char(2) not null,
	
	primary key (ID),
	unique key (movie_id, country_encoding),
	foreign key (`movie_id`) references movies(`tmdb_id`) on delete cascade,
	foreign key (`country_encoding`) references countries(`encoding`) on delete cascade
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

alter table movies
	add constraint FK_Language foreign key(language) references languages(encoding);
