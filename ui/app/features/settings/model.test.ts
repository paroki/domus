import assert from "node:assert/strict";
import test from "node:test";
import {
  buildClearValues,
  buildProviderOptionList,
  buildSettingsDraft,
  CHANNEL_GROUP_NAMES,
  filterModelsByCapability,
  findAgentPersona,
  findChannelGroup,
  getProviderModelsByCapability,
  groupChannelSettings,
  groupSettings,
  indexSettingsByKey,
  isAgentSettingKey,
  isSuperAdmin,
  parseProviderModels,
  SETTINGS_LAUNCHERS,
} from "./model";

test("isSuperAdmin checks user:super-admin role", () => {
  assert.equal(isSuperAdmin(["user:super-admin"]), true);
  assert.equal(isSuperAdmin(["user:user", "ref:parish-admin"]), false);
  assert.equal(isSuperAdmin(undefined), false);
});

test("groupSettings and groupChannelSettings separate agent keys", () => {
  const items = [
    {
      key: "user.telegram.bot_token",
      group: "Telegram Bot",
      label: "Bot Token",
      secret: true,
      value: "••••••••",
      configured: true,
    },
    {
      key: "user.gowa.base_url",
      group: "GOWA (WhatsApp)",
      label: "Base URL",
      secret: false,
      value: "",
      configured: false,
    },
    {
      key: "agent.gabriel.provider_id",
      group: "AI Agent: Gabriel",
      label: "Provider Config",
      secret: false,
      value: "",
      configured: false,
    },
  ];

  assert.equal(isAgentSettingKey("agent.gabriel.model"), true);
  assert.equal(isAgentSettingKey("user.telegram.bot_token"), false);
  assert.equal(groupSettings(items).length, 3);

  const channelGroups = groupChannelSettings(items);
  assert.equal(channelGroups.length, 2);
  assert.equal(channelGroups[0]?.name, "Telegram Bot");
  assert.equal(channelGroups[0]?.configured, true);
  assert.equal(channelGroups[1]?.name, "GOWA (WhatsApp)");
  assert.equal(channelGroups[1]?.configured, false);
});

test("buildClearValues maps every item key to empty string", () => {
  assert.deepEqual(
    buildClearValues([
      {
        key: "user.telegram.bot_token",
        group: "Telegram Bot",
        label: "Bot Token",
        secret: true,
        value: "••••••••",
        configured: true,
      },
    ]),
    { "user.telegram.bot_token": "" },
  );
});

test("parseProviderModels and filterModelsByCapability handle valid and malformed JSON", () => {
  assert.deepEqual(parseProviderModels(undefined), []);
  assert.deepEqual(parseProviderModels("{invalid"), []);
  assert.deepEqual(parseProviderModels('{"not":"array"}'), []);

  const parsed = parseProviderModels(
    JSON.stringify([
      {
        id: "gemini-2.5-flash",
        name: "Gemini 2.5 Flash",
        capabilities: ["chat"],
      },
      {
        id: "text-embedding-004",
        name: "Embedding 004",
        capabilities: ["embedding"],
      },
      { id: "legacy-chat", name: "Legacy Chat" },
    ]),
  );
  assert.equal(parsed.length, 3);

  const chatModels = filterModelsByCapability(parsed, "chat");
  assert.deepEqual(
    chatModels.map((m) => m.id),
    ["gemini-2.5-flash", "legacy-chat"],
  );

  const embModels = filterModelsByCapability(parsed, "embedding");
  assert.deepEqual(
    embModels.map((m) => m.id),
    ["text-embedding-004"],
  );
});

test("buildProviderOptionList and getProviderModelsByCapability work with UUID provider configs", () => {
  const providers = [
    {
      id: "uuid-gem-1",
      name: "Gemini 1",
      provider: "gemini",
      hasApiKey: true,
      models: [
        {
          id: "gemini-2.5-flash",
          name: "Gemini 2.5 Flash",
          capabilities: ["chat"],
        },
        {
          id: "text-embedding-004",
          name: "Embedding 004",
          capabilities: ["embedding"],
        },
      ],
    },
    {
      id: "uuid-ant-1",
      name: "Anthropic 1",
      provider: "anthropic",
      hasApiKey: true,
      models: [
        {
          id: "claude-sonnet-4",
          name: "Claude Sonnet 4",
          capabilities: ["chat"],
        },
      ],
    },
  ];

  const agentOpts = buildProviderOptionList(providers, false);
  assert.equal(agentOpts.length, 2);
  assert.equal(agentOpts[0]?.id, "uuid-gem-1");
  assert.equal(agentOpts[0]?.label, "Gemini 1 (Google AI Studio (Gemini))");

  // Embedding excludes anthropic and appends "none"
  const embOpts = buildProviderOptionList(providers, true);
  assert.equal(embOpts.length, 2);
  assert.equal(embOpts[0]?.id, "uuid-gem-1");
  assert.equal(embOpts[1]?.id, "none");

  assert.deepEqual(
    getProviderModelsByCapability(providers, "uuid-gem-1", "chat").map(
      (m) => m.id,
    ),
    ["gemini-2.5-flash"],
  );
  assert.deepEqual(
    getProviderModelsByCapability(providers, "uuid-gem-1", "embedding").map(
      (m) => m.id,
    ),
    ["text-embedding-004"],
  );
  assert.deepEqual(
    getProviderModelsByCapability(providers, "none", "embedding"),
    [],
  );
});

test("SETTINGS_LAUNCHERS, findAgentPersona, findChannelGroup, and draft helpers work", () => {
  assert.deepEqual(
    SETTINGS_LAUNCHERS.map((item) => item.path),
    [
      "/settings/telegram",
      "/settings/whatsapp",
      "/settings/ai-provider",
      "/settings/ai-agent",
    ],
  );
  assert.equal(findAgentPersona("gabriel")?.name, "Gabriel");
  assert.equal(findAgentPersona("mikael")?.name, "Mikael");
  assert.equal(findAgentPersona("rafael")?.name, "Rafael");
  assert.equal(findAgentPersona("unknown"), undefined);

  const items = [
    {
      key: "user.telegram.bot_token",
      group: CHANNEL_GROUP_NAMES.telegram,
      label: "Bot Token",
      secret: true,
      value: "tok",
      configured: true,
    },
    {
      key: "user.gowa.base_url",
      group: CHANNEL_GROUP_NAMES.whatsapp,
      label: "Base URL",
      secret: false,
      value: "",
      configured: false,
    },
  ];
  assert.equal(
    findChannelGroup(items, CHANNEL_GROUP_NAMES.telegram)?.configured,
    true,
  );
  assert.equal(findChannelGroup(items, "Unknown"), undefined);
  assert.equal(
    indexSettingsByKey(items)["user.telegram.bot_token"]?.label,
    "Bot Token",
  );
  assert.deepEqual(buildSettingsDraft(items), {
    "user.telegram.bot_token": "tok",
    "user.gowa.base_url": "",
  });
});
