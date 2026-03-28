-- U18 jégkorong-világbajnokság (divízió II/A), Kézdivásárhely, 2026-04-12 .. 2026-04-18
-- Forrás: https://sport.szekelyhon.ro/jegkorong/minden-keszen-all-kezdivasarhely-melto-hazigazdaja-szeretne-lenni-az-ifjusagi-hokivebenek
-- Futtatás: psql "$DATABASE_URL" -f backend/migrations/patch_event_u18_hockey_kezdiv_2026.sql

BEGIN;

INSERT INTO settlements (county_id, name, name_ro, slug, type, is_county_seat)
SELECT c.id, 'Kézdivásárhely', 'Târgu Secuiesc', 'kezdivasarhely', 'municípium', false
FROM counties c WHERE c.slug = 'kovaszna'
ON CONFLICT (slug, county_id) DO NOTHING;

INSERT INTO events (location_id, title, description, start_date, start_time, end_date, end_time, event_type, organizer)
SELECT s.id,
    'U18 jégkorong-világbajnokság (II/A divízió) – Kézdivásárhely',
    'Ifjúsági divízió II/A jégkorong-világbajnokság a Deme László Műjégpályán. A romániai U18-as válogatott Nagy-Britannia, Spanyolország, Japán, Kína és Horvátország csapataival mérkőzik meg; minden mérkőzésre díjtalan a belépés, közvetítés: TVR Sport.' || E'\n\n' ||
    'Forrás: https://sport.szekelyhon.ro/jegkorong/minden-keszen-all-kezdivasarhely-melto-hazigazdaja-szeretne-lenni-az-ifjusagi-hokivebenek',
    '2026-04-12',
    NULL,
    '2026-04-18',
    NULL,
    'sports',
    'Román Jégkorongszövetség (FRHG)'
FROM settlements s
JOIN counties c ON s.county_id = c.id
WHERE c.slug = 'kovaszna' AND s.slug = 'kezdivasarhely'
  AND NOT EXISTS (
    SELECT 1 FROM events e
    WHERE e.title = 'U18 jégkorong-világbajnokság (II/A divízió) – Kézdivásárhely'
  );

COMMIT;
