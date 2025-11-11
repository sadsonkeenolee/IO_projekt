CREATE TABLE IF NOT EXISTS `user_credentials` (
  `ID` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` char(64) NOT NULL,
  `password` char(64) NOT NULL,
  `email` char(128) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `credentials_unique` (`username`),
  UNIQUE KEY `credentials_unique_1` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
