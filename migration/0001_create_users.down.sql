DROP TRIGGER IF EXISTS set_timestamp_users ON public.users;
DROP TABLE IF EXISTS public.users;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_language;
DROP FUNCTION IF EXISTS update_updated_at_column();
