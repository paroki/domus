# Changelog

## [0.2.0](https://github.com/paroki/domus/compare/v0.1.0...v0.2.0) (2026-10-11)


### Features

* add active field to parish and diocese with UI and e2e updates ([e8b2334](https://github.com/paroki/domus/commit/e8b23342d2ac9f9f2614500f3d3d877ff96e94bf))
* add territory crud and e2e tests ([84d03d9](https://github.com/paroki/domus/commit/84d03d9f7985571b06804b8da5907ccde1e78196))
* add unit crud and e2e tests ([34823ec](https://github.com/paroki/domus/commit/34823eced147755f4ccb688af1e8752aeeea5169))
* **api:** add /ping endpoint to verify api status :rocket: ([6f41409](https://github.com/paroki/domus/commit/6f4140924184971acfa75ad79a008f604b6d342a))
* **api:** add crud and search endpoints for territory and unit ([64e6462](https://github.com/paroki/domus/commit/64e64621ae90955dba36c52a83467cdc9ca406be))
* **api:** add membership schema ([b43c1d9](https://github.com/paroki/domus/commit/b43c1d9ed77f55051c8876b9331c969a9d0e68d7))
* **api:** add soft delete support ([363c1ee](https://github.com/paroki/domus/commit/363c1ee9148e67b3580fef1289cc5ed680905116))
* **api:** add unit schema ([36df453](https://github.com/paroki/domus/commit/36df4534e8d40b4dad3119154f16724595b2a207))
* **api:** create web response standard and auth integration 🚀 ([1f9a13b](https://github.com/paroki/domus/commit/1f9a13b62223550f29b99c47e36667a1bb14388a))
* **api:** implement crud endpoints for parish and diocese :sparkles: ([226b7e9](https://github.com/paroki/domus/commit/226b7e929a73a862a534fbc15f352278398d9aa4))
* **api:** update openapi generation scripts ([b55238a](https://github.com/paroki/domus/commit/b55238ac2c918da7e7eec0e8b9fcb713f671a6d7))
* **auth:** add social providers integration for Google and GitHub ([51d9086](https://github.com/paroki/domus/commit/51d9086e9e01b27a7cba7dac7fe669cbe361248f))
* **auth:** autocreate superadmin user from env and configure uuid generateId ([e925af3](https://github.com/paroki/domus/commit/e925af363412fe3f871e25e484f11e97ac4b7fb8))
* **dash,e2e:** implement diocese crud pages with tanstack query and e2e test suite ([3f16627](https://github.com/paroki/domus/commit/3f166273e1e9d4e589eb25f9ca599b60f08999f8))
* **dash:** add diocese CRUD with openapi-fetch client ([48010ff](https://github.com/paroki/domus/commit/48010ffcf7ec6e2628c5b8f568b9f65d4460c46d))
* **dash:** add i18n with id/en support ([57d1035](https://github.com/paroki/domus/commit/57d10354d51e2bba2014a31d2a38976d912205b2))
* **dash:** add odoo-style module layout ([d398754](https://github.com/paroki/domus/commit/d3987542b060bacbc17a572032b6e9fe499147db))
* **dash:** add sys module and short module ids ([7875cf6](https://github.com/paroki/domus/commit/7875cf6dafc2d3e37b2cd87a55a34d887e28f1ff))
* **dash:** add under construction placeholder for unbuilt pages ([f155805](https://github.com/paroki/domus/commit/f155805229c3e476d5a54a706f7138f21aafe11d))
* **dash:** bigger base font and colorful module icons ([f25830a](https://github.com/paroki/domus/commit/f25830ad9be96e35bd6974187cc32faa2156b9e1))
* **dash:** implement login page and routing ([5451ccf](https://github.com/paroki/domus/commit/5451ccf740c93b25ca48fbac316c848e8090931f))
* **dash:** improve user menu ([f0377a3](https://github.com/paroki/domus/commit/f0377a34f8902935e262b533523605b30a429ca5))
* **dash:** redirect unauthenticated users to /login ([c1083c2](https://github.com/paroki/domus/commit/c1083c274339b0d4db374677cfacca4a9567e690))
* **e2e:** add login and launcher feature tests ([53ee137](https://github.com/paroki/domus/commit/53ee1372a60daa7321407885c17c07a379bacca1))
* implement parish management with CRUD UI, API attributes, and E2E tests ([ca1e214](https://github.com/paroki/domus/commit/ca1e214dd0c8a2916e9c9c86be7b72977b9a7a94))
* **openapi:** initialize OpenAPI package ([4f036a3](https://github.com/paroki/domus/commit/4f036a351f626f7fba5a3cbe4f5a280ff7d7bb4e))


### Bug Fixes

* **api:** use ephemeral httptest server for jwks mock to prevent port collision ([4e04d8d](https://github.com/paroki/domus/commit/4e04d8dbbce2e5f28d081fd00734326b753ee435))
* **api:** uuid import, API_PORT env, nullable avatar, ErrForbidden envelope ([6409dda](https://github.com/paroki/domus/commit/6409dda4e8d1de20584cd3b49bbf7782ac2d3200))
* **auth:** make authDB lazy and provide env vars in CI test-better-auth ([822183e](https://github.com/paroki/domus/commit/822183e408083e5dc2168285e384a4d491c19fb4))
* **auth:** uncomment request handler in betterAuth setup ([4b39998](https://github.com/paroki/domus/commit/4b39998c46676dba130a8ce4e8286d64e61f5cd7))
* **ci:** fix package name ([3e6d686](https://github.com/paroki/domus/commit/3e6d68682e3b363ece31143b4c9fd3b9d207d1dd))
* **ci:** update biome and turbo config ([1b79150](https://github.com/paroki/domus/commit/1b79150e342c58fa7b2d31f9799a16b73af47f9f))
* **dash:** adjust login max-width and update error message ([3ce71eb](https://github.com/paroki/domus/commit/3ce71ebdc108beb4a0cade9a77ef76897f7d459b))
* **dash:** load user avatar without referrer and fall back to initials ([b2cfc97](https://github.com/paroki/domus/commit/b2cfc9777ad6dc9666fb65cb127056cd4875edc1))
* **e2e:** warm up dash and increase expectPath timeout to prevent initial cold load timeout ([5eff593](https://github.com/paroki/domus/commit/5eff593278f09cc75332131e1793ae0180f318a2))


### Performance Improvements

* **dash:** add vite devtools ([26e1f64](https://github.com/paroki/domus/commit/26e1f64a8c639b922773ac5f4eb1c30b726f829f))


### Dependencies

* **deps:** add @domus/openapi to dependencies ([f569ec1](https://github.com/paroki/domus/commit/f569ec15514b8bfab219199ab8b78e1f7e5100e5))

## 0.1.0 (2026-10-10)


### Features

* **api:** base api project setup ([f7b1d4b](https://github.com/paroki/domus/commit/f7b1d4bc54a4fa25091005d33c6bc9fb599eae6e))
* **auth:** configured base auth server :sparkles: ([85ea7ac](https://github.com/paroki/domus/commit/85ea7acf33e6eeaeb28dea72f990a043b2389916))
* **better-auth:** base configuration :sparkles: ([644f4a4](https://github.com/paroki/domus/commit/644f4a416494b8ce7b09d9c80bc676411d84addd))
* **dash:** base dash project setup ([90c7e64](https://github.com/paroki/domus/commit/90c7e6485bc825adcdd7d220461afaa342c210bd))
