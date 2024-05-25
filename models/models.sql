CREATE TABLE Users (
    ID SERIAL,
    email varchar(255) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    created timestamptz NOT NULL DEFAULT NOW()
);
CREATE TABLE Resources (
    ID SERIAL,
    type varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    path varchar(255) NOT NULL,
    owner_id integer NOT NULL,
    created timestamptz NOT NULL DEFAULT NOW()
);
CREATE TABLE SharedResources (
    owner_id INT NOT NULL,
    fullpath varchar(255) NOT NULL,
    path varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    user_id INT NOT NULL,
    can_write BOOL NOT NULL,
    created timestamptz NOT NULL DEFAULT NOW()
);
CREATE TABLE AuthTokens (
    user_id integer NOT NULL,
    token varchar(255) NOT NULL,
    created timestamptz NOT NULL DEFAULT NOW()
);