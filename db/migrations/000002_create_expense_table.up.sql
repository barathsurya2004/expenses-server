create table if not exists expense_data (
    uuid uuid references user_data(uuid) on delete cascade,
    date_and_time timestamp with time zone not null,
    place varchar(100) not null,
    amount numeric(10, 2) not null,
    currency varchar(3) not null,
    catergory varchar(50) not null
);