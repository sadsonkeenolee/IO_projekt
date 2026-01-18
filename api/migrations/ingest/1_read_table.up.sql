create table if not exists read_table (
	directory varchar(64) not null,
	data_type enum('media', 'book', 'music') not null,
	created_at timestamp default current_timestamp,
	read_at bool default false,
	primary key (directory)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
