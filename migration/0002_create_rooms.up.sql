CREATE TYPE room_type AS ENUM ('small', 'medium', 'large');

CREATE TABLE public.rooms (
    rooms_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    type room_type NOT NULL,
    price_perhour NUMERIC(11, 3) NOT NULL,
    capacity INT NOT NULL,
    img_path TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TRIGGER set_timestamp_rooms
BEFORE UPDATE ON public.rooms
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
