CREATE TABLE IF NOT EXISTS user_identity(
	`ID` bigint unsigned NOT NULL,
	`birthday` date NOT NULL,
	`gender` char(1) NOT NULL,
	`register_date` date NOT NULL,
	`account_status` char(1) NOT NULL,
	
	PRIMARY KEY (`ID`),
	FOREIGN KEY (`ID`) REFERENCES user_credentials(ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
