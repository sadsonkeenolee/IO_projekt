create table if not exists books (
	ID bigint unsigned auto_increment,
	title varchar(128) not null,
	rating float not null,
	isbn char(10) not null,
	isbn13 char(13) not null,
	language varchar(3),
	pages int,
	total_ratings bigint unsigned,
	release_date date null,
	publisher varchar(128),
	primary key (ID),
	unique key (isbn, isbn13)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

create table if not exists authors (
	ID bigint unsigned auto_increment,
	book_id bigint unsigned,
	author varchar(128) not null,
	primary key (ID),
	foreign key (book_id) references books(ID) on delete cascade

)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;
