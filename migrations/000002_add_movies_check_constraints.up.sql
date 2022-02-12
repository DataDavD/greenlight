ALTER TABLE movies
	ADD CONSTRAINT
		movies_run_time CHECK (runtime >= 0);

ALTER TABLE movies
	ADD CONSTRAINT
		movies_year_check CHECK (year BETWEEN 1888 AND DATE_PART('year', NOW()));

ALTER TABLE movies
	ADD CONSTRAINT
		genres_length_check CHECK ( ARRAY_LENGTH(genres, 1) BETWEEN 1 AND 5);
