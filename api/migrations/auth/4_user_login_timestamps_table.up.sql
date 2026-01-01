CREATE TABLE IF NOT EXISTS user_login_timestamps(
	`ID` bigint unsigned auto_increment,
	`user_id` bigint unsigned,
	`token` varchar(255) not null,
	`timestamp` timestamp default current_timestamp,
	
	PRIMARY KEY (`ID`),
	FOREIGN KEY (`user_id`) REFERENCES user_credentials(ID) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
