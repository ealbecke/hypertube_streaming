CREATE TABLE IF NOT EXISTS medias_seen
(
	user_id INT NOT NULL,
	imdb_id VARCHAR(50) NOT NULL,
	CONSTRAINT fk_userid FOREIGN KEY (user_id) REFERENCES users(id)
);