create table if not exists secret (
  id serial primary key,
  hash bytea not null
);