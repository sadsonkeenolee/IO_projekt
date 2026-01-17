CREATE TABLE IF NOT EXISTS `user_credentials` (
  `ID` bigint unsigned auto_increment,
  `username` varchar(32) not null check(char_length(username) >= 1),
  `password` char(60) not null,
  `email` varchar(254) not null,

  PRIMARY KEY (`ID`),
  UNIQUE KEY `credentials_unique` (`username`),
  UNIQUE KEY `credentials_unique_1` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
