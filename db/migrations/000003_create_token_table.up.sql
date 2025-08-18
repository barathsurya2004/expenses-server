create table if not exists token_data (
    uuid uuid references user_data(uuid) on delete cascade primary key,
    token varchar(255) not null unique,
    context varchar(50) not null,
    created_at timestamp with time zone default current_timestamp,
    expires_at timestamp with time zone not null
);