CREATE DATABASE IF NOT EXISTS Hypertube;

USE Hypertube;

CREATE TABLE IF NOT EXISTS languages
(
	id INT PRIMARY KEY NOT NULL AUTO_INCREMENT,
	language VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS providers
(
	id INT PRIMARY KEY NOT NULL AUTO_INCREMENT,
	provider VARCHAR(255) NOT NULL
	
);

CREATE TABLE IF NOT EXISTS users
(
	id INT PRIMARY KEY NOT NULL AUTO_INCREMENT,
	email VARCHAR(255) NOT NULL,
	username VARCHAR(255) NOT NULL,
	firstname VARCHAR(255) NOT NULL,
	lastname VARCHAR(255) NOT NULL,
	password VARCHAR(255),
	language_id INT NOT NULL,
	provider_id INT NOT NULL,
	creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	edit_date TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP,
	CONSTRAINT fk_language_id FOREIGN KEY (language_id) REFERENCES languages(id),
	CONSTRAINT fk_provider_id FOREIGN KEY (provider_id) REFERENCES providers(id)
);

INSERT INTO languages (language) VALUES ("English");
INSERT INTO languages (language) VALUES ("French");
INSERT INTO providers (provider) VALUES ("Local");
INSERT INTO providers (provider) VALUES ("Google");
INSERT INTO providers (provider) VALUES ("42");
INSERT INTO users (email, username, firstname, lastname, password, language_id, provider_id) VALUES ("msourdin@student.42.fr", "msourdin", "Maxime", "Sourdin", "JuliaMonPetitCoquillage", "1", "1");
INSERT INTO users (email, username, firstname, lastname, password, language_id, provider_id) VALUES ("ealbecke@student.42.fr", "ealbecke", "Eliot", "Albecker", "TunningDesIles", "2", "1");
