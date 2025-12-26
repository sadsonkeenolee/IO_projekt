create procedure get_movie_by_id (in movie_id bigint unsigned)
       begin
	  select ID, budget, tmdb_id, language, title, overview,
	    popularity, release_date, revenue, runtime, status,
	    tagline, rating, total_ratings
	  from movies m 
	  where m.tmdb_id=movie_id;
       end;


-- Join every production company, by movie id
create procedure get_production_companies (in movie_id bigint unsigned)
       begin
         select m2c.movie_id, c.company 
         from companies c
         join movie2companies m2c on c.ID = m2c.company_id and m2c.movie_id=movie_id;
       end;

-- Join every production company, by movie id
create procedure get_production_countries (in movie_id bigint unsigned)
       begin
	  select m2c.movie_id, c.country
	  from countries c
	  join movie2countries m2c on c.encoding = m2c.country_encoding and m2c.movie_id=movie_id;
       end;

create procedure get_genres (in movie_id bigint unsigned)
       begin
	  select m2g.movie_id, g.genre
	  from genres g
	  join movie2genres m2g on m2g.genre_id = g.ID and m2g.movie_id=movie_id;
       end;

create procedure get_keywords (in movie_id bigint unsigned)
       begin
	  select m2k.movie_id, k.keyword
	  from keywords k
	  join movie2keywords m2k on m2k.keyword_id = k.ID and m2k.movie_id=movie_id;
       end;

create procedure get_languages (in movie_id bigint unsigned)
       begin
	  select m2l.movie_id, l.language
	  from languages l
	  join movie2languages m2l on m2l.language_encoding = l.encoding and m2l.movie_id=movie_id;
       end;
