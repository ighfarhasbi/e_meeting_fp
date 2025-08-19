CREATE TYPE snack_category AS ENUM ('lunch', 'coffee break');

CREATE TABLE public.snacks (
    snacks_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    price NUMERIC(11, 3) NOT NULL,
    category snack_category NOT NULL
);
