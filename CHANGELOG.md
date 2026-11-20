# Changelog

All notable changes to this project will be documented in this file.

## [v1.0.0-alpha.1](https://github.com/paroki/domus/releases/tag/v1.0.0-alpha.1) (2026-11-20)

### ⚙️ API

#### ✨ Features

- add sentry error reporting ([dbce31c](https://github.com/paroki/domus/commit/dbce31c2a3a51d37a9a22b38623ca388f795e580))
- add me and users endpoints, serve wiring, and openapi spec ([e0716c4](https://github.com/paroki/domus/commit/e0716c4c32d9b614465fb6db21d67f9817f73625))
- add auth middleware, channel resolver, and bootstrap ([7620ff7](https://github.com/paroki/domus/commit/7620ff7b3d00e555872c426dc44ae75f405d9edb))
- add email otp, oauth2, and session auth ([63a0d50](https://github.com/paroki/domus/commit/63a0d50f31a2597e12949f341fd7286da6b10235))
- add user, account, otp, and session ent schemas ([4310503](https://github.com/paroki/domus/commit/4310503fb5f68de63e4850e936414af71b8bbbfe))
- annotate health endpoint and generate OpenAPI spec ([c211d52](https://github.com/paroki/domus/commit/c211d52495106434e26d6b8c2e78ac1d80e5c51b))
- mount mcp endpoint and exclude from SPA fallback ([73746f5](https://github.com/paroki/domus/commit/73746f5e0e1872162978b97d079cea29e681b5fa))
- add MCP server foundation in mcpx ([43e79a5](https://github.com/paroki/domus/commit/43e79a52432743962269bd9e34b4fdede800d287))
- add Casbin RBAC enforcer and Ent adapter in authz ([ff4c80c](https://github.com/paroki/domus/commit/ff4c80ce60241be844bbe6a6d70b31d9fc231b66))
- add Principal context helpers in authx ([b73e18d](https://github.com/paroki/domus/commit/b73e18d7e0d13f40cecfffc9c402e4c05d36ede9))
- add CasbinRule Ent schema and generated client ([d8e914c](https://github.com/paroki/domus/commit/d8e914c0aa12d3bb4eb5804c6a70ef18a4493f2f))
- add domus migrate CLI and serve AutoMigrate ([f4e1dba](https://github.com/paroki/domus/commit/f4e1dbab38feef8e2d7b3bea44afc1239dda7f4c))
- add RunMigrations runner and AutoMigrate config ([d99b301](https://github.com/paroki/domus/commit/d99b30131f699486b97e5e3e6bf5132efffdb8b4))
- add embedded initial SQL migration and atlas.sum ([c67ef15](https://github.com/paroki/domus/commit/c67ef151181b76dad2851ba682539739640a6b3f))
- add database config and OpenDB multi-dialect helper ([873fe12](https://github.com/paroki/domus/commit/873fe12814ab717ed75a1968e284536dccd7f1d6))
- initialize entgo schemas, mixins, and generated client ([631d81d](https://github.com/paroki/domus/commit/631d81d6dec70bafd65eddb375c3daa1a332d563))
- wire httpx ErrorHandler and health envelope ([be562c8](https://github.com/paroki/domus/commit/be562c8cd431b5b4f2f53ad6c5fdd1965e94fb08))
- add httpx envelope, pagination, and error handler ([d67cd49](https://github.com/paroki/domus/commit/d67cd49b8f7f19ecfe73d3c94c23cc6562793212))
- integrate validator/v10 into Fiber StructValidator ([77b9980](https://github.com/paroki/domus/commit/77b99804724aab66ed7badd2c716a7ecfc95435c))

### 🖥️ UI

#### ✨ Features

- add sentry error, tracing and replay reporting ([d68bb1b](https://github.com/paroki/domus/commit/d68bb1bfa54d9ecadd3fb2972d25120dd90ff80b))
- add auth guard, login, me workspace, and users directory ([6dff958](https://github.com/paroki/domus/commit/6dff9581d5f0deed3db4286c3a869800e9da975f))
- add shared API client, QueryClient, and schema types ([6122748](https://github.com/paroki/domus/commit/6122748667e95bfdc9db18688cea969a94e5299f))

### 🔧 Other

#### 🐛 Bug Fixes

- **ci:** prepare release alpha release ([bc154a9](https://github.com/paroki/domus/commit/bc154a9eaed6cb0dd1ac3af79b472bfaf7cb6557))

### 📦 Dependencies

- **deps:** update azure/setup-helm action to v5 (#59) ([2b3afd3](https://github.com/paroki/domus/commit/2b3afd3f41e4799e730fb455c48edaf7c1fe6d91))
- **deps:** update actions/setup-go action to v7 (#58) ([221bb0f](https://github.com/paroki/domus/commit/221bb0f89c76f04ab1a28e03350c432d7e26fbb7))
- **deps:** update module golang.org/x/oauth2 to v0.38.0 (#56) ([80fa215](https://github.com/paroki/domus/commit/80fa215922eb4704594a7318ca3dd79d885e044b))
- **deps:** update module github.com/sethvargo/go-envconfig to v2 (#63) ([465d668](https://github.com/paroki/domus/commit/465d6684538a8637d83b80a0e4ae0587c8f048cb))
- **deps:** update module github.com/casbin/casbin/v2 to v3 (#62) ([50b5647](https://github.com/paroki/domus/commit/50b5647411a43678fe3f1aed43fa16418812d616))
- **deps:** update dependency typescript to v7 (#61) ([10ba153](https://github.com/paroki/domus/commit/10ba153416d689d5a7eb9979891bd6ed20b0bff7))
- **deps:** update dependency @types/node to v26 (#60) ([2504bac](https://github.com/paroki/domus/commit/2504bac555ee25ae14223381c70aeee9f03cd597))
- **deps:** update actions/checkout action to v7 (#57) ([3f06532](https://github.com/paroki/domus/commit/3f06532bcbe72ad8088e59c9a804021867619fcb))
- **deps:** lock file maintenance (#64) ([ec94dc7](https://github.com/paroki/domus/commit/ec94dc7d0dc46977c1c9f3c9793d98c188d1a806))
- **deps:** add react-query, openapi-fetch, and openapi-typescript ([3bb23e3](https://github.com/paroki/domus/commit/3bb23e327361235dc661b7b709cf00dd91b918a0))
- **deps:** include ent codegen CLI checksums in go.sum ([9fa5bd6](https://github.com/paroki/domus/commit/9fa5bd6e45b0dedb5025332afe322b202f089bf8))
- **deps:** add modelcontextprotocol go-sdk dependency ([b06b9df](https://github.com/paroki/domus/commit/b06b9df624815197b708674ff2bdf6703844ca26))
- **deps:** add casbin v2 dependency ([96fc0fa](https://github.com/paroki/domus/commit/96fc0fa2c8a5b48d7521bb51ba1c525654b602bc))
- **deps:** include ent codegen CLI checksums in go.sum ([4fb1404](https://github.com/paroki/domus/commit/4fb14045a8ba99f2a5d46f31f37f1a5489ade7fe))
- **deps:** add ent, uuid, pgx v5, and modernc sqlite ([9b85cbd](https://github.com/paroki/domus/commit/9b85cbd2a3a76552d91a2a05001d13f39b65d19b))
- **deps:** add testify and ubermock v0.6.0 ([882ef93](https://github.com/paroki/domus/commit/882ef93bdf361005f45a67361198566e417daca8))
- **deps:** add go-playground/validator/v10 ([f70877a](https://github.com/paroki/domus/commit/f70877a666864e80fc6d23c226d82e0d2265b726))
- **deps:** add lucide-react and i18next for ui shell ([7c755e8](https://github.com/paroki/domus/commit/7c755e815a0d8146f158d6e75c452ed1bd72b708))
