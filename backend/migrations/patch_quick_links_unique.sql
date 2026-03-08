-- Allow multiple quick links from the same domain (e.g. blog.bogozi.com, bogozi.com, www.bogozi.com)
-- Drop any existing unique constraint that may be too restrictive, then add unique on full URL only.

-- Drop constraint if it exists (handles both custom "unique_quick_link" and default "quick_links_url_key")
ALTER TABLE quick_links DROP CONSTRAINT IF EXISTS unique_quick_link;
ALTER TABLE quick_links DROP CONSTRAINT IF EXISTS quick_links_url_key;

-- Ensure full URL is unique (allows different subdomains: blog.bogozi.com vs bogozi.com)
ALTER TABLE quick_links ADD CONSTRAINT quick_links_url_unique UNIQUE (url);
