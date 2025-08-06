CREATE TABLE IF NOT EXISTS roles (
    id     varchar(26) PRIMARY KEY ,
    name        varchar(10) NOT NULL ,
    access      varchar(40) NOT NULL ,
    created_at  TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    deleted_at  TIMESTAMP
);