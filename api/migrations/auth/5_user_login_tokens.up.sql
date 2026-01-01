create procedure create_user_session (in username varchar(255), in token varchar(255))
begin
insert into user_login_timestamps(user_id, token) values 
	((select ID from user_credentials uc where uc.username=username), token);
end;
