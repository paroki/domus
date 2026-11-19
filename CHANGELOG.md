# Changelog

## 0.1.0-rc.1 (2026-11-19)


### Features

* **api:** ✨ add auth middleware, channel resolver, and bootstrap ([761f65d](https://github.com/paroki/domus/commit/761f65dc4f28a6e1d5fa4c99754df010345d855b))
* **api:** ✨ add database config and OpenDB multi-dialect helper ([7551c98](https://github.com/paroki/domus/commit/7551c989abaac082e7fb54fcc201c69481e7ce4c))
* **api:** ✨ add domus migrate CLI and serve AutoMigrate ([e8474c0](https://github.com/paroki/domus/commit/e8474c005a19447ed0f42b1d518eb460919d1e27))
* **api:** ✨ add email otp, oauth2, and session auth ([9a24ecd](https://github.com/paroki/domus/commit/9a24ecd393becfa4c3c23f4d843129d1c9c1d3e6))
* **api:** ✨ add httpx envelope, pagination, and error handler ([d67cd49](https://github.com/paroki/domus/commit/d67cd49b8f7f19ecfe73d3c94c23cc6562793212))
* **api:** ✨ add MCP server foundation in mcpx ([1679bf8](https://github.com/paroki/domus/commit/1679bf8cc68784dd04ccd926b0c50050ed6fe9fb))
* **api:** ✨ add me and users endpoints, serve wiring, and openapi spec ([b9bb7c2](https://github.com/paroki/domus/commit/b9bb7c2097f37ebef86fd21ff478087febeb10a3))
* **api:** ✨ add RunMigrations runner and AutoMigrate config ([1450ec3](https://github.com/paroki/domus/commit/1450ec3b3891defcfab90ce78641b2c3fcc511ae))
* **api:** ✨ add sentry error reporting ([9a81420](https://github.com/paroki/domus/commit/9a8142054f3f7b7e7193877f49426a6a1c1352a4))
* **api:** ✨ add user, account, otp, and session ent schemas ([8c0c966](https://github.com/paroki/domus/commit/8c0c96617334af56e782bfa67062a0794e7403fe))
* **api:** ✨ annotate health endpoint and generate OpenAPI spec ([9736439](https://github.com/paroki/domus/commit/9736439b5bdac6616b89708df0f02ed88e514fd0))
* **api:** ✨ integrate validator/v10 into Fiber StructValidator ([77b9980](https://github.com/paroki/domus/commit/77b99804724aab66ed7badd2c716a7ecfc95435c))
* **api:** ✨ mount mcp endpoint and exclude from SPA fallback ([b8655f3](https://github.com/paroki/domus/commit/b8655f3cd5f731bd66c64a8ef37638e7db8ebffd))
* **api:** ✨ wire httpx ErrorHandler and health envelope ([be562c8](https://github.com/paroki/domus/commit/be562c8cd431b5b4f2f53ad6c5fdd1965e94fb08))
* **auth:** ✨ add Casbin RBAC enforcer and Ent adapter in authz ([00c4773](https://github.com/paroki/domus/commit/00c47733fc0f4d471df8991398d521168074f278))
* **auth:** ✨ add Principal context helpers in authx ([6397870](https://github.com/paroki/domus/commit/6397870b6050b10ae7bc34563c6dc20fe6800506))
* **db:** ✨ add CasbinRule Ent schema and generated client ([f7e9fe9](https://github.com/paroki/domus/commit/f7e9fe928c746ac3a9321d34bcf2e50be8080abe))
* **db:** ✨ add embedded initial SQL migration and atlas.sum ([656b6ed](https://github.com/paroki/domus/commit/656b6ede34af1afb6bb5100f231a8e63e82fbb7c))
* **db:** ✨ initialize entgo schemas, mixins, and generated client ([89e60b2](https://github.com/paroki/domus/commit/89e60b250f72126b29cb15dda6d01a8719733687))
* **ui:** ✨ add auth guard, login, me workspace, and users directory ([dbaf4d9](https://github.com/paroki/domus/commit/dbaf4d9970900e8457f993fe4e2657e3b89ac197))
* **ui:** ✨ add sentry error, tracing and replay reporting ([743f3e0](https://github.com/paroki/domus/commit/743f3e095f12d702a8220e2932bf95857fa760bc))
* **ui:** ✨ add shared API client, QueryClient, and schema types ([020bf15](https://github.com/paroki/domus/commit/020bf156243be0438af5e71732687d31427647f5))


### Dependencies

* **deps:** 🔧 add casbin v2 dependency ([7bb36d6](https://github.com/paroki/domus/commit/7bb36d6cc2ed937012788e4604724db163318568))
* **deps:** 🔧 add ent, uuid, pgx v5, and modernc sqlite ([9b85cbd](https://github.com/paroki/domus/commit/9b85cbd2a3a76552d91a2a05001d13f39b65d19b))
* **deps:** 🔧 add go-playground/validator/v10 ([f70877a](https://github.com/paroki/domus/commit/f70877a666864e80fc6d23c226d82e0d2265b726))
* **deps:** 🔧 add lucide-react and i18next for ui shell ([7c755e8](https://github.com/paroki/domus/commit/7c755e815a0d8146f158d6e75c452ed1bd72b708))
* **deps:** 🔧 add modelcontextprotocol go-sdk dependency ([96a37c8](https://github.com/paroki/domus/commit/96a37c8e94268f483e9293b0d5ddf4aaa559da36))
* **deps:** 🔧 add react-query, openapi-fetch, and openapi-typescript ([36efe2d](https://github.com/paroki/domus/commit/36efe2d8fbbd176191db0d336ae5d6e3cafd4144))
* **deps:** 🔧 add testify and ubermock v0.6.0 ([882ef93](https://github.com/paroki/domus/commit/882ef93bdf361005f45a67361198566e417daca8))
* **deps:** 🔧 include ent codegen CLI checksums in go.sum ([0d2716d](https://github.com/paroki/domus/commit/0d2716d0129509a228a1213ae5190311fd2171de))
* **deps:** 🔧 include ent codegen CLI checksums in go.sum ([f5a12d8](https://github.com/paroki/domus/commit/f5a12d84effee77bd4286fa7340e3a21154d54ab))


### Miscellaneous Chores

* **release:** 🔧 release 0.1.0-rc.1 ([29b2def](https://github.com/paroki/domus/commit/29b2defbdde2d912f55d8ecb97c23296542c1127))
