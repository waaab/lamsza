-- Drop legacy tables that are no longer used by the application.
-- Safe to run multiple times thanks to IF EXISTS and CASCADE.

DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS service_categories CASCADE;

