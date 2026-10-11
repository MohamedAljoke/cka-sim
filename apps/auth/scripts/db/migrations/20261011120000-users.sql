-- +migrate Up
create table users (
    id            uuid        primary key default gen_random_uuid(),
    email         text        not null,
    password_hash text        not null,
    created_at    timestamptz not null default now()
);

create unique index users_email_key on users (lower(email));

-- +migrate Down
drop table users;
