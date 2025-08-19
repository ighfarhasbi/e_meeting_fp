CREATE TABLE public.detail_transaction (
    detail_tx_id UUID PRIMARY KEY,
    tx_id UUID NOT NULL,
    rooms_id INT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    participants INT NOT NULL,
    snacks_id INT,
    sub_total_snacks NUMERIC(11, 3),
    sub_total_price_room NUMERIC(11, 3) NOT NULL,
    price_snack_perpack NUMERIC(11, 3),
    price_room_perhour NUMERIC(11, 3) NOT NULL,
    CONSTRAINT detail_transaction_tx_id_fkey FOREIGN KEY (tx_id) REFERENCES public.transactions(tx_id),
    CONSTRAINT detail_transaction_rooms_id_fkey FOREIGN KEY (rooms_id) REFERENCES public.rooms(rooms_id),
    CONSTRAINT detail_transaction_snacks_id_fkey FOREIGN KEY (snacks_id) REFERENCES public.snacks(snacks_id)
);
