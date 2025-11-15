CREATE TABLE IF NOT EXISTS read_table (
	`directory` varchar(64) not null,
	`data_type` enum('media', 'book', 'music') not null,
	`is_read` bool default false,
	`created_at` timestamp default current_timestamp,
	`read_at` timestamp null,
	primary key (directory)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
