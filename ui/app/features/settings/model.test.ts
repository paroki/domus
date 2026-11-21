import assert from "node:assert/strict";
import test from "node:test";
import {
  buildClearValues,
  buildProviderOptionList,
  filterModelsByCapability,
  groupChannelSettings,
  groupSettings,
  isAgentSettingKey,
  isProviderConfigured,
  isSuperAdmin,
  parseProviderModels,
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
      key: "agent.provider.gemini.api_key",
      group: "AI Provider: Gemini",
      label: "Gemini API Key",
      secret: true,
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

test("isProviderConfigured and buildProviderOptionList reflect persisted and draft state", () => {
  const itemsByKey = {
    "agent.provider.gemini.api_key": {
      key: "agent.provider.gemini.api_key",
      configured: true,
      value: "••••••••",
    },
    "agent.provider.openai.api_key": {
      key: "agent.provider.openai.api_key",
      configured: false,
      value: "",
    },
  };
  const draft = {
    "agent.provider.openai.api_key": "sk-new-draft",
  };

  assert.equal(isProviderConfigured(itemsByKey, {}, "gemini"), true);
  assert.equal(isProviderConfigured(itemsByKey, {}, "openai"), false);
  assert.equal(isProviderConfigured(itemsByKey, draft, "openai"), true);

  const opts = buildProviderOptionList(itemsByKey, draft, true);
  assert.equal(opts.length, 4);
  assert.equal(opts[0]?.configured, true);
  assert.equal(opts[1]?.configured, true);
  assert.equal(opts[2]?.configured, false);
  assert.equal(opts[3]?.id, "none");
});
