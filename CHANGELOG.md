# Changelog

All notable changes to this project will be documented in this file.

## [v1.0.0-alpha.6](https://github.com/paroki/domus/compare/v1.0.0-alpha.5...v1.0.0-alpha.6) (2026-11-24)

### 📦 Dependencies

- **deps:** remove unused a2ui and markdown packages ([173d37d](https://github.com/paroki/domus/commit/173d37dd3aa68d1728928cfced85f648987844b0))
- **deps:** update dependency atlas to v1.4.0 (#164) ([82558d4](https://github.com/paroki/domus/commit/82558d496e70f689280b857a3fd0050c36da9e70))
- **deps:** update dependency typescript to v7 (#166) ([c32dd3c](https://github.com/paroki/domus/commit/c32dd3c22dd015ba3c007690e3d1ef755e41ab68))

## [v1.0.0-alpha.5](https://github.com/paroki/domus/compare/v1.0.0-alpha.4...v1.0.0-alpha.5) (2026-11-23)

### ⚙️ API

#### ✨ Features

- add a2ui v0.9 web prompt instructions and block sanitizer ([f09ac9d](https://github.com/paroki/domus/commit/f09ac9d0d4becf59887426b56e89e5cbb3c90233))
- add org_create MCP tool for parish admins ([4311ed3](https://github.com/paroki/domus/commit/4311ed31b251ec1f9e46ef03abfa72f5f64f3be3))
- handoff directly to specialist on delegation ([3892998](https://github.com/paroki/domus/commit/38929986b35e344d25918997e972faa2a8c30a7c))
- support continuing interrupted agent chat turns ([83cc776](https://github.com/paroki/domus/commit/83cc776d3d8d4e91912f594bb6d1f8cb9eef325e))
- allow emoji and unicode characters in agent prompts ([8e225eb](https://github.com/paroki/domus/commit/8e225ebcd2671d0b76b72d4a56b54a9b271d95fa))
- support multi-config LLM providers with UUID binding ([9e71463](https://github.com/paroki/domus/commit/9e71463720f05ca5fc3ff676e437395d86f655fe))

#### 🐛 Bug Fixes

- inject Casbin user roles into agent system prompt ([31c8111](https://github.com/paroki/domus/commit/31c81119cdc5f84951e5ed9c223b0122c475a93b))
- instruct Gabriel to delegate silently ([71b1275](https://github.com/paroki/domus/commit/71b127578ab228ad65e15f8a1315b6ac7a9f07b3))
- add ErrProviderOverloaded sentinel for upstream 5xx responses ([8a10293](https://github.com/paroki/domus/commit/8a10293304b69ee2c393decef546e5ed03fa717e))
- normalize Gemini type arrays and separate provider error sentinels ([bf11e65](https://github.com/paroki/domus/commit/bf11e65cd17036650b5ab6281a456c7b4f3a85a0))
- disable Fiber WriteTimeout and add SSE keepalive for agent chat ([f5d5383](https://github.com/paroki/domus/commit/f5d538357eb80697aa273456a1a85e21c2a50c1b))
- log agent SSE stream errors and preserve provider details ([93c4690](https://github.com/paroki/domus/commit/93c469072a23e628129dd569856b7c2c3009ad8a))
- capture all 5xx errors in sentry middleware ([0b5b4d9](https://github.com/paroki/domus/commit/0b5b4d911ce17a74d74c95cd10405b1c24f26a44))
- return 400 for unavailable AI provider ([d5e84f7](https://github.com/paroki/domus/commit/d5e84f7b8cd00ca265e7999487aa86b6045c5da1))
- skip unavailable errors in sentry middleware ([0d53b43](https://github.com/paroki/domus/commit/0d53b43042bbcfdfa2165e90b59c70cd9b4c7ba2))
- skip 4xx errors in sentry middleware ([edcbe84](https://github.com/paroki/domus/commit/edcbe84a995caae86b3247ee501fe8f8492934e5))
- allow empty baseUrl on provider update ([fc0ec7c](https://github.com/paroki/domus/commit/fc0ec7c654b695cb699c09731d5d0141f54b66d1))
- skip gemma thinkingConfig and strip unsupported schema keys ([040d25f](https://github.com/paroki/domus/commit/040d25f6639f22e3e137e012b5dd1f41bd9f0568))
- require only bot token for telegram config ([8beffd0](https://github.com/paroki/domus/commit/8beffd0d42e8a9fd5d1eb7eb30840fa5a89fc937))
- prevent stale SPA cache and missing asset HTML fallback ([967ab10](https://github.com/paroki/domus/commit/967ab10a015cb8234bd198f8e3f042c75c3e9bba))
- add missing agent_providers migration ([d7c964a](https://github.com/paroki/domus/commit/d7c964a5e52a62ee1e8d9b0abd1e5c513babbff6))
- require postgres extensions on startup ([7033a6b](https://github.com/paroki/domus/commit/7033a6b0ad636ce9b029dbf9e4a9095326d3b0b4))

### 🖥️ UI

#### ✨ Features

- add a2ui theme tokens and refine agent chat and settings ([7eed469](https://github.com/paroki/domus/commit/7eed4690b2c20efd28f8026c80a468a9f0084b8a))
- add a2ui surface renderer and agent streaming fixes ([605b658](https://github.com/paroki/domus/commit/605b658d917d034887f4ddfb1b7d3e3a599b885e))
- add continue from last support on interrupted turns ([1e800dc](https://github.com/paroki/domus/commit/1e800dce8d2fe686e782bcd46f7d7101d238ba06))
- render markdown in agent chat timeline ([454c292](https://github.com/paroki/domus/commit/454c2924ec30aeecbcef12a61edf157ef64a3884))
- add retry button to agent chat error banner ([ae6c612](https://github.com/paroki/domus/commit/ae6c612da2fb2aa284d674a8cb254aa6dc4a2d16))
- show Gabriel and Menu tabs on root home ([f2a0b3d](https://github.com/paroki/domus/commit/f2a0b3de5b0c12505e204350f659a96ccd52e472))
- add Update API Key toggle button on provider edit form ([2cb61ea](https://github.com/paroki/domus/commit/2cb61ea083920784beb37827e175372f31c83ade))
- manage multiple LLM provider configs by UUID ([fce60e7](https://github.com/paroki/domus/commit/fce60e7daa4b00e13b73560f0af5309780522dd8))

#### 🐛 Bug Fixes

- add AGENT_PROVIDER_OVERLOADED error message ([2336654](https://github.com/paroki/domus/commit/233665420492b37f14a32c2df6e060f5fd8a3854))
- map distinct agent provider error codes ([b95c5cf](https://github.com/paroki/domus/commit/b95c5cf4d71bafae7187b1591a86a64cc8f1e964))
- surface provider error details from agent SSE stream ([a04686e](https://github.com/paroki/domus/commit/a04686e5a3889925d97592a291731b8b620e2db1))
- use Button and canvas theme tokens instead of text-bg ([a27a9b4](https://github.com/paroki/domus/commit/a27a9b401841176e2b9e86ba718aaea35ae85665))

### 🔧 Other

#### 🐛 Bug Fixes

- **tools:** restore postgres extensions after e2e schema reset ([7725c25](https://github.com/paroki/domus/commit/7725c2506ed3fd99733d7a7d06735ee609a1c2f1))

### 📦 Dependencies

- **deps:** add a2ui react and web_core packages ([56e7cda](https://github.com/paroki/domus/commit/56e7cda29859fcef6762444e63991904b7f5fd9c))
- **deps:** add firebase genkit go v1.14.0 ([3f8723a](https://github.com/paroki/domus/commit/3f8723a0cb5ee4c15ac5c2fb2766d1161d79a1fc))
- **deps:** add react-markdown and remark-gfm for agent chat ([4ebcbd1](https://github.com/paroki/domus/commit/4ebcbd179327722dd6c3315a17ef3442ea4cec20))
- **deps:** remove modernc sqlite ([0c02ceb](https://github.com/paroki/domus/commit/0c02cebbb30edfe17467394411a1d046aef5a7b1))

## [v1.0.0-alpha.4](https://github.com/paroki/domus/compare/v1.0.0-alpha.3...v1.0.0-alpha.4) (2026-11-22)

### ⚙️ API

#### ✨ Features

- add telegram and gowa whatsapp webhook adapters for agent ([c1030ca](https://github.com/paroki/domus/commit/c1030cae6041b3222c7f0069d21684e36a6d33e6))
- add user gowa settings and ref parish code lookup ([a14b7cf](https://github.com/paroki/domus/commit/a14b7cfbd11c97197d4b2a142cee9b8f18a9a4a0))
- extend gowa and telegram clients for agent webhooks ([33b3b82](https://github.com/paroki/domus/commit/33b3b82222ec2d1e47af3d65a0df00bb52ccd3ba))
- add agent http/sse api, permissions, and serve wiring ([7720e21](https://github.com/paroki/domus/commit/7720e21e8f36d96fac3d495146bcbbf595f2d512))
- add memory tools and multi-agent react orchestrator ([2a3a104](https://github.com/paroki/domus/commit/2a3a10437d00a81a9233d34ade074bfc320af604))
- add in-process mcp bridge and user channel lookup ([a95e004](https://github.com/paroki/domus/commit/a95e0044a0c8ace25913a5b2ef0c38dc4ad642b5))
- add agent settings registration and model sync ([b52e9c0](https://github.com/paroki/domus/commit/b52e9c074414453ff18c097ac5161b83edad79fa))
- add multi-provider llm and embedding drivers ([cc5df05](https://github.com/paroki/domus/commit/cc5df05cc9680d713491cce730b42f9454af87c8))
- add embedded agent persona prompts ([f999620](https://github.com/paroki/domus/commit/f99962073bd5dfee5e359d6573d4119f6c2e821e))
- encrypt secret settings and support defaults ([43721b6](https://github.com/paroki/domus/commit/43721b6504de62d8d4d79ce9eeb2ab8ebca07f88))
- add agent domain contract and 3-tier memory repo ([c2ce72c](https://github.com/paroki/domus/commit/c2ce72c60924a15c0c4beedbc861e3c9489b52cf))
- add errx KindUnavailable and httpx 503 mapping ([7d1d672](https://github.com/paroki/domus/commit/7d1d6725c06622ea6c52f7496a04dfceaf1c2c34))
- add agent session, message, and memory schemas ([15db6cc](https://github.com/paroki/domus/commit/15db6cccdba72c5b82ba0200439698a735e8a2f4))
- add gender and birthDate to user schema and me endpoint ([27cbfb0](https://github.com/paroki/domus/commit/27cbfb0666a6f074a0f9d369829a62d239df8995))

#### 🐛 Bug Fixes

- drop overall llm client timeout for sse streams ([a1456d2](https://github.com/paroki/domus/commit/a1456d219f946a1d9d5fbfe47d9fb019ebed8fb5))
- block nested specialist delegation ([6478c5b](https://github.com/paroki/domus/commit/6478c5b1fe6521c985b381049af602389fee158f))
- sanitize agent history and cap tool results ([c82fcb5](https://github.com/paroki/domus/commit/c82fcb5f309bd51cdcd7014e7e88150be2091204))
- match whatsapp group mentions against bot number ([033517c](https://github.com/paroki/domus/commit/033517ccc5a95871219653085eb3153272fcf7bc))
- run agent webhooks async with dedup and fix mentions ([42aa6ee](https://github.com/paroki/domus/commit/42aa6eebba1fd1578fb0bf5a0ebbf5dedce953ac))
- cancel agent sse turn on disconnect and timeout ([e50460d](https://github.com/paroki/domus/commit/e50460d8e369aefd9778a0be7d5d6e95ac05ce64))
- add lowercase json tags on ProviderModelItem ([8ba1ada](https://github.com/paroki/domus/commit/8ba1ada04f4ff8af235a6703abf0e22b09b9c4aa))

### 🖥️ UI

#### ✨ Features

- add ai assistant workspace, memory drawer, and sse chat ([5a56515](https://github.com/paroki/domus/commit/5a56515339aae9bf8e03b9cabcc6bfb8875b32d8))
- add llm providers and agent config tabs to settings ([1b19225](https://github.com/paroki/domus/commit/1b1922519d6c55f88a38c98232bd720e968363e1))
- add gender and birthDate fields to onboarding and profile ([0d192ff](https://github.com/paroki/domus/commit/0d192ff78bd970cf111112081f4bcb0db842e153))

## [v1.0.0-alpha.3](https://github.com/paroki/domus/compare/v1.0.0-alpha.2...v1.0.0-alpha.3) (2026-11-21)

### ⚙️ API

#### ✨ Features

- wire telegram and gowa settings hot-reload ([ea14e0b](https://github.com/paroki/domus/commit/ea14e0b667d3199c6aa47e517f976437106b47b3))
- add platform setting store and api ([1f92516](https://github.com/paroki/domus/commit/1f92516927a90bf51cfa9aaeebf8d1e22637d680))
- replace meta schema with setting ([f09f419](https://github.com/paroki/domus/commit/f09f419963eb6835b2c87186786348f2f544691c))
- require APP_SECRET at startup and initialize cipher ([9cb65af](https://github.com/paroki/domus/commit/9cb65af37c8dff47b6fef03c4a27f44863594ed6))
- add platform crypt package for AES-256-GCM and blind index ([db8ac32](https://github.com/paroki/domus/commit/db8ac329e4106d2b90453d8936555a997f22a674))
- add role catalog, roleDetails, and org:admin resolution ([c824f1a](https://github.com/paroki/domus/commit/c824f1aba2833d1f9eb101ac49c45cf74fc35eb4))
- add onboarding flag, linking, and Telegram webhook ([7951c2e](https://github.com/paroki/domus/commit/7951c2e024a4644fa0252ec51891090b73034dc4))
- add multi-channel OTP auth and startup env validation ([896eea0](https://github.com/paroki/domus/commit/896eea010da1329dcd73ed82bea7e5b33d93bf70))
- add GOWA WhatsApp REST client and JID normalizer ([560253c](https://github.com/paroki/domus/commit/560253c4080e4766c4b07a160ad695a993605a5a))
- add Telegram Bot HTTP client and webhook verifier ([fcf5783](https://github.com/paroki/domus/commit/fcf5783a2ce62f2e133093547edf335cdd4bd1c7))

#### 🐛 Bug Fixes

- remove trailing punctuation in setting error strings ([3572409](https://github.com/paroki/domus/commit/35724099426410ccccb76da9cae3bd24fa272b85))
- set user avatar from oauth when empty ([ae9138c](https://github.com/paroki/domus/commit/ae9138c1c967627b06c83c559899fd5e8d327a7b))
- normalize timestamptz to timestamp for sqlite migrations ([8b9929f](https://github.com/paroki/domus/commit/8b9929ff3fcbe85e14a5854d8b1ec39f6262ec0c))

### 🖥️ UI

#### ✨ Features

- add super admin settings page ([77e8714](https://github.com/paroki/domus/commit/77e8714a201f307aa95cff7ea68a67caf4128e8c))
- centralize parish and org member management in users page ([51d4e95](https://github.com/paroki/domus/commit/51d4e95764bbd000786ece6cb9c7b556c78f21ac))
- add 3-step onboarding wizard and RequireAuth gate ([bad514e](https://github.com/paroki/domus/commit/bad514e5edeacd524a00a8024e3f60a118d4e489))
- support multi-channel OTP login and channel linking ([250cc42](https://github.com/paroki/domus/commit/250cc429a4297bfcdec3634683b75ee906aa4728))

#### 🐛 Bug Fixes

- restore Church import and format ([dbdb3ba](https://github.com/paroki/domus/commit/dbdb3ba0b6c2a2d2a34f61c4e4101fd894ad2ae4))
- add shared unwrapVoid for 204 No Content mutations ([45d0b46](https://github.com/paroki/domus/commit/45d0b469a351346a9a38f57db1582a41f761b5ba))
- sync verified WhatsApp phone in onboarding state ([da90606](https://github.com/paroki/domus/commit/da90606e0de0e3de63a7daabcd90619e331f0261))

### 📦 Dependencies

- **deps:** pin ui typescript to ~5.9.3 for openapi-typescript ([8a1bf68](https://github.com/paroki/domus/commit/8a1bf68899a047a297556787768a17d96649737b))
- **deps:** add uqr for client-side Telegram QR SVG rendering ([e02abea](https://github.com/paroki/domus/commit/e02abea843a29eefe719aaac4eb734b7c8ce78ca))

## [v1.0.0-alpha.2](https://github.com/paroki/domus/compare/v1.0.0-alpha.1...v1.0.0-alpha.2) (2026-11-21)

### ⚙️ API

#### ✨ Features

- expose event MCP tools and integration tests ([1d0fe8d](https://github.com/paroki/domus/commit/1d0fe8d56ce5ed99418cad3c838bb0d5f3746c26))
- implement event RSVP upsert, cutoff, summary, and roster API ([f03c412](https://github.com/paroki/domus/commit/f03c412c30fdf143e98b644cf184cfd097d30915))
- implement event CRUD, visibility rules, and Casbin RBAC ([7662db1](https://github.com/paroki/domus/commit/7662db1103edb39631910e49854479871a340636))
- add Event and EventRSVP Ent schemas, migration, and in-use guards ([3e2350a](https://github.com/paroki/domus/commit/3e2350a115aac4148e055c7a1fdc70f2350ee7ac))
- add org mcp tools and tests ([d2e1466](https://github.com/paroki/domus/commit/d2e146673486584fbf80cf740feee8b70adda9c1))
- add org invites and guest auth onboarding ([4ea2784](https://github.com/paroki/domus/commit/4ea2784130c1239bb80c1940adb99a95e059fdaf))
- implement org crud, members, casbin, and facade ([682d210](https://github.com/paroki/domus/commit/682d210bd5f0e4603f75bbb648f1fe9bb5fec728))
- add org ent schemas and migration ([cef0a34](https://github.com/paroki/domus/commit/cef0a34e32546a13f8cec03c09949eb9d70a04cb))
- add stdlib SMTP mailer and wire email OTP delivery ([24700a4](https://github.com/paroki/domus/commit/24700a48a3818ab754b7f6a9af279982a2442748))
- add zone ward lookup and period CRUD ([3925cd9](https://github.com/paroki/domus/commit/3925cd94f5450092d086787b0f4e11263e268fe4))
- add ref feature for diocese parish and members ([36da2b0](https://github.com/paroki/domus/commit/36da2b06ab30d6849702a2ca9daa41fe7c9fb909))
- include parish roles in user session and profile ([5555272](https://github.com/paroki/domus/commit/55552727a780cf99c15e56addbd5a79bce4b10cc))
- add ref Ent schemas and Atlas migration ([16b727b](https://github.com/paroki/domus/commit/16b727b7937f4754b432db479bbb38f4a2f100bc))

#### 🐛 Bug Fixes

- fix org migration indexes, invite maxUses, member and period guards ([ecdfe3c](https://github.com/paroki/domus/commit/ecdfe3c45fade0a6dfa1f22097a0c0e1fd6d4ec0))
- enforce parish membership on session switch and header ([4eb20c4](https://github.com/paroki/domus/commit/4eb20c430dd0d99d82789f64000e776f84bdedc9))
- enforce global lookup uniqueness, member name, and sort order ([d602f13](https://github.com/paroki/domus/commit/d602f139ea662b9524003704ac82b7f200dd9069))

### 🖥️ UI

#### ✨ Features

- implement event feed, detail workspace, RSVP widget, and roster ([6c25ec7](https://github.com/paroki/domus/commit/6c25ec7778b68630fcd590d7c9b57bded5a7c5b6))
- implement org directory, roster, invites, and join ui ([c0a8da0](https://github.com/paroki/domus/commit/c0a8da06b256f099e539ccc44f5d50a6f8094190))
- add territory, lookup, and period management pages ([3131d60](https://github.com/paroki/domus/commit/3131d601d9ce1455f923162672c3c96619ce90b4))
- add parish switcher, settings grid, and parish pages ([c4af258](https://github.com/paroki/domus/commit/c4af2589c18f039113749a75c460013e4e0671be))

#### 🐛 Bug Fixes

- fix org period filter, inline edit, invite revoke and join flash ([dcc9435](https://github.com/paroki/domus/commit/dcc9435d3ca0c70b84196de732ff8b5031658afb))
- clear optional ref fields and share session hooks ([cf6f83c](https://github.com/paroki/domus/commit/cf6f83cc4d8060a9fff538ab87317aae41a7b7d7))

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
