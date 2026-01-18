create procedure if not exists push_events(
in p_token varchar(512), 
in p_event enum('like', 'dislike', 'playlist'), 
in p_type enum('book', 'tv', 'movie'),
in p_item_id bigint unsigned
)
begin
	insert into user_events(token, event, type, item_id, timestamp) 
	values (p_token, p_event, p_type, p_item_id, current_timestamp);
end;

create procedure if not exists pull_events(
in p_token varchar(512),
in p_event enum('like', 'dislike', 'playlist'))
begin
	select item_id, event, type 
	from user_events
	where (
	token=p_token and 
	event=p_event
	);
end;
