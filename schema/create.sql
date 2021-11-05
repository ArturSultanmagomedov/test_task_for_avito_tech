create table if not exists users
(
    id      serial primary key,
    user_id int   not null unique,
    balance float not null
);