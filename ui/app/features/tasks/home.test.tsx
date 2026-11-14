import assert from "node:assert/strict";
import test from "node:test";
import { renderToString } from "react-dom/server";
import Home, { meta } from "./home";

test("meta return title and description", () => {
  // ponytail: empty obj cast; route typegen not needed for static meta
  const args = {} as unknown as Parameters<typeof meta>[0];
  const tags = meta(args);
  assert.deepEqual(tags, [
    { title: "Domus" },
    { name: "description", content: "Parish Administration Digitalized" },
  ]);
});

test("Home render heading", () => {
  const html = renderToString(<Home />);
  assert.ok(html.includes("Domus: Parish Administration Digitalized"));
});
