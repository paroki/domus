import assert from "node:assert/strict";
import test from "node:test";
import {
  buildClearValues,
  buildSettingsDraft,
  CHANNEL_GROUP_NAMES,
  filterModelsByCapability,
  findChannelGroup,
  groupSettings,
  indexSettingsByKey,
  isSuperAdmin,
  parseProviderModels,
  SETTINGS_LAUNCHERS,
} from "./model";

test("isSuperAdmin checks user:super-admin role", () => {
  assert.equal(isSuperAdmin(["user:super-admin"]), true);
  assert.equal(isSuperAdmin(["user:user", "ref:parish-admin"]), false);
  assert.equal(isSuperAdmin(undefined), false);
});

test("groupSettings and findChannelGroup group channel settings", () => {
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
  ];

  const groups = groupSettings(items);
  assert.equal(groups.length, 2);
  assert.equal(groups[0]?.name, "Telegram Bot");
  assert.equal(groups[0]?.configured, true);
  assert.equal(groups[1]?.name, "GOWA (WhatsApp)");
  assert.equal(groups[1]?.configured, false);
  assert.equal(
    findChannelGroup(items, CHANNEL_GROUP_NAMES.telegram)?.name,
    "Telegram Bot",
  );
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

test("SETTINGS_LAUNCHERS, indexSettingsByKey, and buildSettingsDraft work", () => {
  assert.deepEqual(
    SETTINGS_LAUNCHERS.map((item) => item.id),
    ["telegram", "whatsapp", "aiProvider"],
  );
  const items = [
    { key: "user.telegram.bot_token", value: "tok", configured: true },
  ];
  assert.equal(
    indexSettingsByKey(items)["user.telegram.bot_token"]?.value,
    "tok",
  );
  assert.deepEqual(buildSettingsDraft(items), {
    "user.telegram.bot_token": "tok",
  });
});
