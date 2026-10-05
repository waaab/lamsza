# Admin extraction (local)

Main Lámsza admin UI/API moved to the independent repo [`waaab/lamsza-admin`](https://github.com/waaab/lamsza-admin).

| | |
|--|--|
| Local admin | http://localhost:5173/ (API `:3000`) |
| Main app | http://localhost:5174/ (API `:3001`) — **no** `/admin` route |
| Shared DB | Same Postgres as main; migrations stay in this repo |
| Prod | `admin.lamsza.com` = **Phase 2** |

Szótár and Játszótér admin UIs are out of scope for this extraction.
