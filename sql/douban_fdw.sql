CREATE EXTENSION douban_fdw;
CREATE SERVER doubansv FOREIGN DATA WRAPPER douban_fdw;
CREATE FOREIGN TABLE top250(rating REAL, title TEXT, genres VARCHAR(64), casts VARCHAR(256), collectcount INTEGER, originname TEXT, directors VARCHAR(256), year VARCHAR(32)) SERVER doubansv OPTIONS(rank_name 'top250');

SELECT count(1) FROM top250;

SELECT rating, title, CAST(year AS INT) FROM top250 WHERE rating > 9.2 ORDER BY title DESC;

SELECT rating, title, CAST(year AS INT) FROM top250 WHERE CAST(year AS INT) = 1994 ORDER BY title;

DROP FOREIGN TABLE top250;
DROP SERVER doubansv;
DROP EXTENSION douban_fdw CASCADE;