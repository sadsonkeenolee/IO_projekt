CREATE TABLE IF NOT EXISTS `user_credentials` (
  `ID` bigint unsigned auto_increment,
  `username` char(64) not null,
  `password` char(64) not null,
  `email` varchar(128) not null,
  
  PRIMARY KEY (`ID`),
  UNIQUE KEY `credentials_unique` (`username`),
  UNIQUE KEY `credentials_unique_1` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
