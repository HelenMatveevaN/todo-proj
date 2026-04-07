create table tasks (
	id 			SERIAL	primary key,
	title 		text	not null,
	content		text,
	is_done		boolean	default false,
	created_at	timestamp	default now()
);