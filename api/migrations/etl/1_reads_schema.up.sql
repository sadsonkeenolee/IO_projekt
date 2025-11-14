CREATE TABLE IF NOT EXISTS schema_reads (
	`directory` varchar(64) not null,
	`data_type` enum('media', 'book', 'music') not null,
	`is_read` bool default false,
	`created_at` timestamp default current_timestamp,
	`read_at` timestamp null,
	primary key (directory)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO schema_reads(directory, data_type) VALUES 
('goodreads-books-data', 'book'),
('books-data', 'book'),
('tmdb-movies-data', 'media'),
('movies-data', 'media')
