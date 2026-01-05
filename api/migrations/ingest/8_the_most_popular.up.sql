create materialized view top_100_shows as
select ID, budget, tmdb_id, language, title, overview,
	popularity, release_date, revenue, runtime, status,
	tagline, rating, total_ratings
from movies
order by popularity desc
limit 100;

create materialized view top_100_books as
select ID, title, isbn, isbn13, language, pages, release_date, publisher, 
	rating, total_ratings
from books 
order by total_ratings desc, rating desc
limit 100;

create materialized view default_shows_recommendation as
select ID, budget, tmdb_id, language, title, overview,
	popularity, release_date, revenue, runtime, status,
	tagline, rating, total_ratings
from movies
order by popularity desc
limit 25;

create materialized view default_books_recommendation as
select ID, title, isbn, isbn13, language, pages, release_date, publisher, 
	rating, total_ratings
from books 
order by total_ratings desc, rating desc
limit 25;

