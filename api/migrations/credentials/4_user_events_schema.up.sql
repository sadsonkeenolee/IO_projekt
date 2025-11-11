CREATE TABLE IF NOT EXISTS user_events(
	`ID` bigint unsigned NOT NULL,
	`user_id` bigint unsigned NOT NULL,
	-- Nie wiem jeszcze jak eventy beda ogarniane
	`event` VARCHAR(255),
	`timestamp` TIMESTAMP,
	
	PRIMARY KEY (`ID`),
	FOREIGN KEY (`user_id`) REFERENCES user_credentials(`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
