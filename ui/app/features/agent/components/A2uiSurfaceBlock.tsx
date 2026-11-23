import { A2uiSurface, type ReactCatalogComponent } from "@a2ui/react/v0_9";
import {
  MessageProcessor,
  type ProcessableMessage,
  type SurfaceModel,
} from "@a2ui/web_core/v0_9";
import {
  basicCatalog,
  injectBasicCatalogStyles,
} from "@a2ui/web_core/v0_9/basic_catalog";
import { useEffect, useRef, useState } from "react";

let stylesInjected = false;

function ensureA2uiStyles() {
  if (stylesInjected || typeof document === "undefined") return;
  injectBasicCatalogStyles();
  stylesInjected = true;
}

export function A2uiSurfaceBlock({
  messages,
  onAction,
}: {
  messages: Record<string, unknown>[];
  onAction?: (actionPrompt: string) => void;
}) {
  const onActionRef = useRef(onAction);
  onActionRef.current = onAction;

  const messagesKey = JSON.stringify(messages);
  const messagesRef = useRef(messages);
  messagesRef.current = messages;

  const [surfaces, setSurfaces] = useState<
    SurfaceModel<ReactCatalogComponent>[]
  >([]);
  const [renderFailed, setRenderFailed] = useState(false);

  useEffect(() => {
    ensureA2uiStyles();
    const processor = new MessageProcessor<ReactCatalogComponent>(
      [basicCatalog],
      async (action) => {
        const context = action.context ?? {};
        const hasContext = Object.keys(context).length > 0;
        const summary = hasContext
          ? `Aksi UI "${action.name}" dengan data: ${JSON.stringify(context)}`
          : `Aksi UI "${action.name}"`;
        onActionRef.current?.(summary);
      },
    );

    const sync = () =>
      setSurfaces(Array.from(processor.model.surfacesMap.values()));
    const createdSub = processor.onSurfaceCreated(sync);
    const deletedSub = processor.onSurfaceDeleted(sync);

    setRenderFailed(false);
    try {
      processor.processMessages(
        messagesRef.current as unknown as ProcessableMessage[],
      );
      sync();
    } catch (err) {
      console.warn("A2UI render failed:", err);
      setRenderFailed(true);
    }

    return () => {
      createdSub.unsubscribe();
      deletedSub.unsubscribe();
    };
  }, [messagesKey]);

  if (renderFailed) {
    return (
      <div className="rounded-xl border border-line bg-surface px-3 py-2 text-xs text-muted">
        ⚠ Tampilan interaktif tidak dapat dirender.
      </div>
    );
  }

  if (surfaces.length === 0) return null;

  return (
    <div className="flex flex-col gap-2 rounded-xl border border-line bg-surface p-3">
      {surfaces.map((surface) => (
        <A2uiSurface key={surface.id} surface={surface} />
      ))}
    </div>
  );
}
