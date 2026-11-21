import assert from "node:assert/strict";
import test from "node:test";
import { buildClearValues, groupSettings, isSuperAdmin } from "./model";

test("isSuperAdmin checks user:super-admin role", () => {
  assert.equal(isSuperAdmin(["user:super-admin"]), true);
  assert.equal(isSuperAdmin(["user:user", "ref:parish-admin"]), false);
  assert.equal(isSuperAdmin(undefined), false);
});

test("groupSettings groups items in order and computes configured status", () => {
  const groups = groupSettings([
    {
      key: "user.telegram.bot_token",
      group: "Telegram Bot",
      label: "Bot Token",
      secret: true,
      value: "••••••••",
      configured: true,
    },
    {
      key: "user.telegram.bot_username",
      group: "Telegram Bot",
      label: "Bot Username",
      secret: false,
      value: "domus_bot",
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
  ]);

  assert.equal(groups.length, 2);
  assert.equal(groups[0]?.name, "Telegram Bot");
  assert.equal(groups[0]?.items.length, 2);
  assert.equal(groups[0]?.configured, true);
  assert.equal(groups[1]?.name, "GOWA (WhatsApp)");
  assert.equal(groups[1]?.configured, false);
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
      {
        key: "user.telegram.bot_username",
        group: "Telegram Bot",
        label: "Bot Username",
        secret: false,
        value: "domus_bot",
        configured: true,
      },
    ]),
    {
      "user.telegram.bot_token": "",
      "user.telegram.bot_username": "",
    },
  );
});
