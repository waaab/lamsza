-- 2026-10-07: the daily mondás lives only in Szótár (its `proverbs`, managed in
-- the admin app's /dictionary; OPEN_ITEMS, Mondások). Lámsza's home page reads
-- it from Szótár's /api/proverbs/today and nothing reads this table any more.
-- Every row in it was test data (owner's decision); a dump of dev's rows is in
-- ~/.cache/lamsza-network/backups/2026-10-07-mondasok/. Apply in production only
-- after the code without it is deployed.
DROP TABLE IF EXISTS mondasok;
