-- +migrate Up
alter table docs.chunks add column domain text;

-- +migrate Down
alter table docs.chunks drop column domain;
