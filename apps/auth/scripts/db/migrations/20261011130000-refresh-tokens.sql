-- +migrate Up
create table refresh_tokens (
    id         bigint      generated always as identity primary key,
    family_id  uuid        not null,
    account_id uuid        not null references users (id) on delete cascade,
    token_hash bytea       not null unique,
    expires_at timestamptz not null,
    used_at    timestamptz,
    revoked_at timestamptz,
    created_at timestamptz not null default now()
);

create index refresh_tokens_family_id on refresh_tokens (family_id);
create index refresh_tokens_account_id on refresh_tokens (account_id);

-- +migrate Down
drop table refresh_tokens;
