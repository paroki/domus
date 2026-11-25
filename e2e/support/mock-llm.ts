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

function lastUserMessage(messages: ChatMessage[]): string {
  for (let i = messages.length - 1; i >= 0; i--) {
    if (messages[i]?.role === "user") {
      return String(messages[i]?.content ?? "");
    }
  }
  return "";
}

function hasToolResultInCurrentTurn(messages: ChatMessage[]): boolean {
  let lastUserIdx = -1;
  for (let i = messages.length - 1; i >= 0; i--) {
    if (messages[i]?.role === "user") {
      lastUserIdx = i;
      break;
    }
  }
  for (let i = lastUserIdx + 1; i < messages.length; i++) {
    if (messages[i]?.role === "tool") {
      return true;
    }
  }
  return false;
}

function pickToolCall(
  tools: ChatTool[],
  userText: string,
): { name: string; arguments: string } | null {
  const names = tools
    .map((t) => t.function?.name ?? "")
    .filter((n) => n.length > 0);
  if (names.length === 0) return null;

  // Gabriel supervisor routing via adk.NewAgentTool
  if (names.includes("mikael") || names.includes("rafael")) {
    if (/kegiatan|jadwal|acara/i.test(userText) && names.includes("mikael")) {
      return {
        name: "mikael",
        arguments: JSON.stringify({ request: userText }),
      };
    }
    if (
      /organisasi|paroki|wilayah|lingkungan/i.test(userText) &&
      names.includes("rafael")
    ) {
      return {
        name: "rafael",
        arguments: JSON.stringify({ request: userText }),
      };
    }
    return null;
  }

  // Sub-agent leaf MCP tools
  if (names.includes("event_list")) {
    return { name: "event_list", arguments: "{}" };
  }
  if (names.includes("org_list")) {
    return { name: "org_list", arguments: "{}" };
  }
  if (names.includes("ref_get_paroki")) {
    return { name: "ref_get_paroki", arguments: "{}" };
  }
  return null;
}

function writeSseChunk(
  res: http.ServerResponse,
  delta: Record<string, unknown>,
  finishReason?: string,
): void {
  const choice: Record<string, unknown> = { index: 0, delta };
  if (finishReason) {
    choice.finish_reason = finishReason;
  }
  res.write(
    `data: ${JSON.stringify({
      id: "chatcmpl-e2e",
      object: "chat.completion.chunk",
      created: 1,
      model: "gpt-4.1-mini",
      choices: [choice],
    })}\n\n`,
  );
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

      if (body.stream !== true) {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(
          JSON.stringify({
            id: "chatcmpl-json-1",
            object: "chat.completion",
            created: 1,
            model: "gpt-4.1-mini",
            choices: [
              {
                index: 0,
                message: { role: "assistant", content: "[]" },
                finish_reason: "stop",
              },
            ],
          }),
        );
        return;
      }

      const userText = lastUserMessage(messages);
      const hasToolResult = hasToolResultInCurrentTurn(messages);

      res.writeHead(200, {
        "Content-Type": "text/event-stream",
        "Cache-Control": "no-cache",
      });

      if (userText.includes("[SLOW_STREAM]")) {
        writeSseChunk(res, {
          role: "assistant",
          content: "Memulai proses analisis paroki... ",
        });
        const timer = setInterval(() => {
          if (res.writableEnded) {
            clearInterval(timer);
            return;
          }
          writeSseChunk(res, { content: "memeriksa data... " });
        }, 120);
        timer.unref?.();
        res.on("close", () => {
          clearInterval(timer);
          if (!res.writableEnded) {
            res.end();
          }
        });
        return;
      }

      const chosenTool = !hasToolResult ? pickToolCall(tools, userText) : null;
      if (chosenTool) {
        writeSseChunk(res, {
          role: "assistant",
          tool_calls: [
            {
              index: 0,
              id: `call_e2e_${chosenTool.name}`,
              type: "function",
              function: {
                name: chosenTool.name,
                arguments: chosenTool.arguments,
              },
            },
          ],
        });
        writeSseChunk(res, {}, "tool_calls");
        res.write("data: [DONE]\n\n");
        res.end();
        return;
      }

      if (hasToolResult) {
        await new Promise<void>((resolve) => {
          const t = setTimeout(resolve, 1200);
          req.on("close", () => {
            clearTimeout(t);
            resolve();
          });
        });
        if (res.writableEnded) return;
      }

      const summaryText = hasToolResult
        ? "Berdasarkan data paroki, informasi kegiatan dan organisasi sudah tersedia."
        : "Halo! Ada yang bisa Gabriel bantu untuk keperluan paroki Anda?";

      writeSseChunk(
        res,
        {
          role: "assistant",
          content: summaryText,
        },
        "stop",
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
  s.closeAllConnections?.();
  await new Promise<void>((resolve, reject) => {
    s.close((err) => (err ? reject(err) : resolve()));
  });
}
