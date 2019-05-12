CREATE EXTENSION douban_fdw;
CREATE SERVER doubansv FOREIGN DATA WRAPPER douban_fdw;
CREATE FOREIGN TABLE top250(rating REAL, title TEXT, genres VARCHAR(64), casts VARCHAR(256), collectcount INTEGER, originname TEXT, directors VARCHAR(256), year VARCHAR(32)) SERVER doubansv OPTIONS(rank_name 'top250');

SELECT count(1) FROM top250;

SELECT originname, title FROM top250 WHERE CAST(year AS INT) = 1994 ORDER BY rating DESC, title;

SELECT title, year FROM top250 WHERE rating NOT IN (SELECT rating FROM top250 WHERE year::int > 2000) ORDER BY rating DESC, title;

EXPLAIN SELECT rating, title, year FROM top250 WHERE rating NOT IN (SELECT rating FROM top250 WHERE year::int > 2000) ORDER BY rating DESC;

DROP FOREIGN TABLE top250;
DROP SERVER doubansv;
DROP EXTENSION douban_fdw CASCADE;