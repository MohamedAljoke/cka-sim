-- +migrate Up

-- Development fixture only, a second account for checking labs are per user. Credentials: dev2@cka-sim.local / cka-sim-dev
insert into users (id, email, password_hash)
values (
    '019dc5a0-0000-7000-8000-000000000002',
    'dev2@cka-sim.local',
    '$argon2id$v=19$m=19456,t=2,p=1$JB6KcCd88YHvtJ/hEYjzGQ$li5I+j7q/yrp73jKQaBMBDvcmuORSkUlM0b9DkcMKKs'
)
on conflict do nothing;

-- +migrate Down
delete from users where id = '019dc5a0-0000-7000-8000-000000000002';
