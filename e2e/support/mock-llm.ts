import http from "node:http";
import type { AddressInfo } from "node:net";

let server: http.Server | null = null;
let baseUrl = "";

interface ChatTool {
  function?: {
    name?: string;
  };
}

interface ChatMessage {
  role?: string;
  content?: string;
}

function readJsonBody(
  req: http.IncomingMessage,
): Promise<Record<string, unknown>> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    req.on("data", (chunk: Buffer) => chunks.push(chunk));
    req.on("end", () => {
      const raw = Buffer.concat(chunks).toString("utf8").trim();
      if (!raw) {
        resolve({});
        return;
      }
      try {
        resolve(JSON.parse(raw) as Record<string, unknown>);
      } catch (err) {
        reject(err);
      }
    });
    req.on("error", reject);
  });
}

function pickToolCall(
  tools: ChatTool[],
): { name: string; arguments: string } | null {
  const names = tools
    .map((t) => t.function?.name ?? "")
    .filter((n) => n.length > 0);
  if (names.length === 0) return null;

  if (names.includes("event_list")) {
    return { name: "event_list", arguments: "{}" };
  }
  if (names.includes("org_list")) {
    return { name: "org_list", arguments: "{}" };
  }
  return { name: names[0], arguments: "{}" };
}

export async function startMockLLMServer(): Promise<string> {
  if (server) return baseUrl;

  server = http.createServer(async (req, res) => {
    const url = req.url ?? "";

    if (req.method === "GET" && url.endsWith("/models")) {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(
        JSON.stringify({
          data: [
            { id: "gpt-4.1-mini", name: "GPT-4.1 Mini" },
            { id: "gpt-4.1", name: "GPT-4.1" },
            { id: "text-embedding-3-small", name: "Text Embedding 3 Small" },
          ],
        }),
      );
      return;
    }

    if (req.method === "POST" && url.endsWith("/embeddings")) {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(
        JSON.stringify({
          data: [{ embedding: Array.from({ length: 768 }, () => 0.01) }],
        }),
      );
      return;
    }

    if (req.method === "POST" && url.endsWith("/chat/completions")) {
      const body = await readJsonBody(req);
      const messages = Array.isArray(body.messages)
        ? (body.messages as ChatMessage[])
        : [];
      const tools = Array.isArray(body.tools) ? (body.tools as ChatTool[]) : [];
      const hasToolResult = messages.some((m) => m.role === "tool");

      res.writeHead(200, {
        "Content-Type": "text/event-stream",
        "Cache-Control": "no-cache",
      });

      const chosenTool = !hasToolResult ? pickToolCall(tools) : null;
      if (chosenTool) {
        res.write(
          `data: ${JSON.stringify({
            choices: [
              {
                delta: {
                  reasoning_content:
                    "Menganalisis permintaan umat dan memeriksa data paroki...",
                },
              },
            ],
          })}\n\n`,
        );
        res.write(
          `data: ${JSON.stringify({
            choices: [
              {
                delta: {
                  tool_calls: [
                    {
                      index: 0,
                      id: "call_e2e_1",
                      type: "function",
                      function: {
                        name: chosenTool.name,
                        arguments: chosenTool.arguments,
                      },
                    },
                  ],
                },
              },
            ],
          })}\n\n`,
        );
        res.write(
          `data: ${JSON.stringify({
            choices: [{ delta: {}, finish_reason: "tool_calls" }],
          })}\n\n`,
        );
        res.write("data: [DONE]\n\n");
        res.end();
        return;
      }

      res.write(
        `data: ${JSON.stringify({
          choices: [
            {
              delta: {
                content: "Berdasarkan data paroki, ",
              },
            },
          ],
        })}\n\n`,
      );
      res.write(
        `data: ${JSON.stringify({
          choices: [
            {
              delta: {
                content: "informasi kegiatan dan organisasi sudah tersedia.",
              },
              finish_reason: "stop",
            },
          ],
        })}\n\n`,
      );
      res.write("data: [DONE]\n\n");
      res.end();
      return;
    }

    res.writeHead(404);
    res.end("Not Found");
  });

  await new Promise<void>((resolve) => {
    server?.listen(0, "127.0.0.1", () => resolve());
  });
  const addr = server.address() as AddressInfo;
  baseUrl = `http://127.0.0.1:${addr.port}/v1`;
  return baseUrl;
}

export async function stopMockLLMServer(): Promise<void> {
  if (!server) return;
  const s = server;
  server = null;
  baseUrl = "";
  await new Promise<void>((resolve, reject) => {
    s.close((err) => (err ? reject(err) : resolve()));
  });
}
