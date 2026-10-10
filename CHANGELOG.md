# Changelog

## [0.2.0](https://github.com/paroki/domus/compare/v0.1.0...v0.2.0) (2026-10-10)


### Features

* **api:** add /ping endpoint to verify api status :rocket: ([6f41409](https://github.com/paroki/domus/commit/6f4140924184971acfa75ad79a008f604b6d342a))
* **api:** create web response standard and auth integration 🚀 ([1f9a13b](https://github.com/paroki/domus/commit/1f9a13b62223550f29b99c47e36667a1bb14388a))
* **api:** implement crud endpoints for parish and diocese :sparkles: ([226b7e9](https://github.com/paroki/domus/commit/226b7e929a73a862a534fbc15f352278398d9aa4))
* **auth:** add social providers integration for Google and GitHub ([51d9086](https://github.com/paroki/domus/commit/51d9086e9e01b27a7cba7dac7fe669cbe361248f))
* **dash:** add i18n with id/en support ([57d1035](https://github.com/paroki/domus/commit/57d10354d51e2bba2014a31d2a38976d912205b2))
* **dash:** add odoo-style module layout ([d398754](https://github.com/paroki/domus/commit/d3987542b060bacbc17a572032b6e9fe499147db))
* **dash:** add under construction placeholder for unbuilt pages ([f155805](https://github.com/paroki/domus/commit/f155805229c3e476d5a54a706f7138f21aafe11d))
* **dash:** bigger base font and colorful module icons ([f25830a](https://github.com/paroki/domus/commit/f25830ad9be96e35bd6974187cc32faa2156b9e1))
* **dash:** implement login page and routing ([5451ccf](https://github.com/paroki/domus/commit/5451ccf740c93b25ca48fbac316c848e8090931f))
* **dash:** improve user menu ([f0377a3](https://github.com/paroki/domus/commit/f0377a34f8902935e262b533523605b30a429ca5))
* **dash:** redirect unauthenticated users to /login ([c1083c2](https://github.com/paroki/domus/commit/c1083c274339b0d4db374677cfacca4a9567e690))


### Bug Fixes

* **api:** uuid import, API_PORT env, nullable avatar, ErrForbidden envelope ([6409dda](https://github.com/paroki/domus/commit/6409dda4e8d1de20584cd3b49bbf7782ac2d3200))
* **auth:** uncomment request handler in betterAuth setup ([4b39998](https://github.com/paroki/domus/commit/4b39998c46676dba130a8ce4e8286d64e61f5cd7))
* **ci:** update biome and turbo config ([1b79150](https://github.com/paroki/domus/commit/1b79150e342c58fa7b2d31f9799a16b73af47f9f))
* **dash:** adjust login max-width and update error message ([3ce71eb](https://github.com/paroki/domus/commit/3ce71ebdc108beb4a0cade9a77ef76897f7d459b))
* **dash:** load user avatar without referrer and fall back to initials ([b2cfc97](https://github.com/paroki/domus/commit/b2cfc9777ad6dc9666fb65cb127056cd4875edc1))


### Performance Improvements

* **dash:** add vite devtools ([26e1f64](https://github.com/paroki/domus/commit/26e1f64a8c639b922773ac5f4eb1c30b726f829f))

## 0.1.0 (2026-10-10)


### Features

* **api:** base api project setup ([f7b1d4b](https://github.com/paroki/domus/commit/f7b1d4bc54a4fa25091005d33c6bc9fb599eae6e))
* **auth:** configured base auth server :sparkles: ([85ea7ac](https://github.com/paroki/domus/commit/85ea7acf33e6eeaeb28dea72f990a043b2389916))
* **better-auth:** base configuration :sparkles: ([644f4a4](https://github.com/paroki/domus/commit/644f4a416494b8ce7b09d9c80bc676411d84addd))
* **dash:** base dash project setup ([90c7e64](https://github.com/paroki/domus/commit/90c7e6485bc825adcdd7d220461afaa342c210bd))
