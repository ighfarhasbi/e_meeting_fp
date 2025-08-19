CREATE TYPE tx_status_enum AS ENUM ('booked', 'paid', 'canceled');

CREATE TABLE public.transactions (
    tx_id UUID PRIMARY KEY,
    users_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    no_hp VARCHAR(15) NOT NULL,
    company VARCHAR(255) NOT NULL,
    status tx_status_enum NOT NULL DEFAULT 'booked',
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    canceled_at TIMESTAMPTZ,
    total NUMERIC(11, 3) NOT NULL,
    CONSTRAINT transactions_users_id_fkey FOREIGN KEY (users_id) REFERENCES public.users(users_id)
);

CREATE TRIGGER set_timestamp_transactions
BEFORE UPDATE ON public.transactions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
