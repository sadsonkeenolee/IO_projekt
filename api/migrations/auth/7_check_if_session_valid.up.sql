create function if not exists check_if_session_is_valid(
f_token varchar(255)
)
returns boolean
deterministic
reads sql data 
begin 
	declare last_session timestamp;
	select coalesce(MAX(`timestamp`),'2038-01-19 03:14:07') into last_session
	from user_login_timestamps
	where token=f_token;
	return timestampdiff(second, current_timestamp, last_session) < 3600;
end;
