CREATE TABLE IF NOT EXISTS user_identity(
	`ID` bigint unsigned auto_increment,
	`birthday` timestamp not null,
	`gender` enum('M', 'F', 'N') not null,
	`register_date` timestamp default current_timestamp,
	`account_status` enum('active', 'inactive', 'banned') default 'inactive',
	
	PRIMARY KEY (`ID`),
	FOREIGN KEY (`ID`) REFERENCES user_credentials(ID) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
