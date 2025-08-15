-- Write your migrate up statements here
create table users if not exists(
    id serial primary key,
    username text unique,
    email text unique not null,
    passhash text not null,
    create_at timestamptz not null default now(),
    update_at timestamptz not null default now()
);
---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
drop table if exists users;