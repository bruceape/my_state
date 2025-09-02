BEGIN;
-- your SQL here
CREATE TABLE IF NOT EXISTS weather (
    city varchar(80),
    date date
);

COMMIT;
