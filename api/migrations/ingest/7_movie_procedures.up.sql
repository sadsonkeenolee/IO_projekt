create procedure if not exists get_movie_by_id (in movie_id bigint unsigned)
       begin
	  select ID, budget, tmdb_id, language, title, overview,
	    popularity, release_date, revenue, runtime, status,
	    tagline, rating, total_ratings
	  from movies m 
	  where m.tmdb_id=movie_id;
       end;

create procedure if not exists find_movie_id(in q_title varchar(255))
       begin
	  select tmdb_id 
	  from movies m
	  where m.title like concat('%', q_title, '%')
	  order by m.popularity desc
	  limit 1;
       end;

create procedure if not exists get_production_companies (in movie_id bigint unsigned)
       begin
         select m2c.movie_id, c.company 
         from companies c
         join movie2companies m2c on c.ID = m2c.company_id and m2c.movie_id=movie_id;
       end;

create procedure if not exists get_production_countries (in movie_id bigint unsigned)
       begin
	  select m2c.movie_id, c.country
	  from countries c
	  join movie2countries m2c on c.encoding = m2c.country_encoding and m2c.movie_id=movie_id;
       end;

create procedure if not exists get_genres (in movie_id bigint unsigned)
       begin
	  select m2g.movie_id, g.genre
	  from genres g
	  join movie2genres m2g on m2g.genre_id = g.ID and m2g.movie_id=movie_id;
       end;

create procedure if not exists get_keywords (in movie_id bigint unsigned)
       begin
	  select m2k.movie_id, k.keyword
	  from keywords k
	  join movie2keywords m2k on m2k.keyword_id = k.ID and m2k.movie_id=movie_id;
       end;

create procedure if not exists get_languages (in movie_id bigint unsigned)
       begin
	  select m2l.movie_id, l.language
	  from languages l
	  join movie2languages m2l on m2l.language_encoding = l.encoding and m2l.movie_id=movie_id;
       end;

create procedure if not exists get_book_by_id (in book_id bigint unsigned)
       begin
	  select 
	    ID, title, isbn, isbn13, language, 
	    pages, release_date, publisher, rating, total_ratings
	  from books b 
	  where b.ID=book_id;
       end;

create procedure if not exists get_book_authors(in book_id bigint unsigned)
      begin
	 select a.ID, a.author
	 from authors a
	 where a.book_id like concat('%', book_id, '%');
      end;

create procedure if not exists find_book_id (in title varchar(255))
       begin
	  select ID	  
	  from books b 
	  where b.title like concat('%', title, '%');
       end;
