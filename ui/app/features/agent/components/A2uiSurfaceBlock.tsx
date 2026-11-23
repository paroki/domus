import { A2uiSurface, type ReactCatalogComponent } from "@a2ui/react/v0_9";
import { MessageProcessor, type ProcessableMessage } from "@a2ui/web_core/v0_9";
import {
  basicCatalog,
  injectBasicCatalogStyles,
} from "@a2ui/web_core/v0_9/basic_catalog";
import { useEffect, useMemo, useRef, useState } from "react";

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

  const processor = useMemo(() => {
    ensureA2uiStyles();
    return new MessageProcessor<ReactCatalogComponent>(
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
  }, []);

  const [surfaces, setSurfaces] = useState(() =>
    Array.from(processor.model.surfacesMap.values()),
  );

  useEffect(() => {
    const sync = () =>
      setSurfaces(Array.from(processor.model.surfacesMap.values()));
    const createdSub = processor.onSurfaceCreated(sync);
    const deletedSub = processor.onSurfaceDeleted(sync);
    try {
      processor.processMessages(messages as unknown as ProcessableMessage[]);
      sync();
    } catch {
      // Ignore invalid A2UI envelopes safely; prose remains visible.
    }
    return () => {
      createdSub.unsubscribe();
      deletedSub.unsubscribe();
    };
  }, [processor, messages]);

  if (surfaces.length === 0) return null;

  return (
    <div className="flex flex-col gap-2 rounded-xl border border-line bg-surface p-3">
      {surfaces.map((surface) => (
        <A2uiSurface key={surface.id} surface={surface} />
      ))}
    </div>
  );
}
