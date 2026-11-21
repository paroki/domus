import { useMemo } from "react";
import { encode } from "uqr";

interface QrCodeSvgProps {
  value: string;
  size?: number;
  label?: string;
  className?: string;
}

export function QrCodeSvg({
  value,
  size = 192,
  label = "Kode QR",
  className,
}: QrCodeSvgProps) {
  const { total, path } = useMemo(() => {
    const qr = encode(value, { ecc: "M", border: 0 });
    const margin = 4;
    const cmds: string[] = [];

    for (let y = 0; y < qr.size; y++) {
      const row = qr.data[y];
      if (!row) continue;
      for (let x = 0; x < qr.size; x++) {
        if (row[x]) {
          cmds.push(`M${x + margin} ${y + margin}h1v1h-1z`);
        }
      }
    }

    return {
      total: qr.size + margin * 2,
      path: cmds.join(""),
    };
  }, [value]);

  return (
    <svg
      width={size}
      height={size}
      viewBox={`0 0 ${total} ${total}`}
      shapeRendering="crispEdges"
      role="img"
      aria-label={label}
      className={className}
    >
      <rect width={total} height={total} fill="#ffffff" />
      <path d={path} fill="#000000" />
    </svg>
  );
}
