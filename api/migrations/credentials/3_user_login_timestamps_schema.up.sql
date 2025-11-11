CREATE TABLE IF NOT EXISTS user_login_timestamps(
	`ID` bigint unsigned NOT NULL,
	`timestamp` timestamp NOT NULL,
	
	PRIMARY KEY (`ID`, `timestamp`),
	FOREIGN KEY (`ID`) REFERENCES user_credentials(ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
