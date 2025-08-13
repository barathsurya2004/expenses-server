create table if not exists user_data (
    uuid uuid primary key,
    username varchar(50) not null unique,
    email varchar(100) not null unique,
    password_hash bytea not null,
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp,
    first_name varchar(50),
    last_name varchar(50)
);