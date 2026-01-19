create table if not exists user_events(
	`ID` bigint unsigned auto_increment,
	`token` varchar(512) not null,
	`event` enum('like', 'dislike', 'playlist', 'unplaylist') not null,
	`type` enum('book', 'tv', 'movie') not null,
	`item_id` bigint unsigned not null,
	`timestamp` timestamp default current_timestamp,

	primary key (`ID`),
	foreign key (`token`) references user_login_timestamps(`token`) on delete cascade
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
