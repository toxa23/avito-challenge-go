create table if not exists "user" (
	id bigserial not null primary key,
	username varchar null,
	balance bigint not null default 0  CHECK (balance >= 0),
	created_at timestamp not null default now()
);

create table if not exists "transaction" (
	id bigserial not null primary key,
	user_id bigint not null,
	amount bigint not null default 0,
    "details" varchar,
	created_at timestamp not null default now()
);

create index if not exists ix_transaction_user_id on
"transaction"
	using btree (user_id, created_at);

alter table "transaction" drop constraint if exists transaction_user_id_fkey;
alter table "transaction" add constraint transaction_user_id_fkey foreign key (user_id) references "user"(id);

insert into "user" (balance) values(100); -- create sample user 1
insert into "user" (balance) values(100); -- create sample user 2
