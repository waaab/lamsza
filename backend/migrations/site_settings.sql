-- Site settings for weather cache and provider selection
CREATE TABLE IF NOT EXISTS site_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO site_settings (key, value) VALUES
  ('weather_cache_ttl_minutes', '15'),
  ('weather_cache_version', '1'),
  ('quick_links_version', '1'),
  ('weather_icon_style', 'emoji'),
  ('weather_active_users_estimate', '10000'),
  ('weather_provider_default', 'open_meteo'),
  ('weather_provider_open_meteo_enabled', 'true'),
  ('weather_provider_weatherapi_enabled', 'true'),
  ('weather_provider_openweathermap_enabled', 'true')
ON CONFLICT (key) DO NOTHING;
