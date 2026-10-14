-- +migrate Up
create extension if not exists vector;

create schema docs;

create table docs.chunks (
    id        text primary key,
    url       text not null,
    title     text not null,
    text      text not null,
    hash      text not null,
    embedding vector(768) not null,
    tsv       tsvector generated always as (to_tsvector('english', title || ' ' || text)) stored
);

create index chunks_embedding on docs.chunks using hnsw (embedding vector_cosine_ops);
create index chunks_tsv on docs.chunks using gin (tsv);

create table docs.meta (
    key   text primary key,
    value text not null
);

-- +migrate Down
drop schema docs cascade;
