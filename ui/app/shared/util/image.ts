export const DEFAULT_MAX_DIMENSION = 1600;
export const DEFAULT_MAX_BYTES = 2 * 1024 * 1024;
export const DEFAULT_IMAGE_QUALITY = 0.82;

export class ImageTooLargeError extends Error {
  readonly sizeBytes: number;
  readonly maxBytes: number;

  constructor(sizeBytes: number, maxBytes: number = DEFAULT_MAX_BYTES) {
    const sizeMb = (sizeBytes / (1024 * 1024)).toFixed(2);
    const maxMb = (maxBytes / (1024 * 1024)).toFixed(0);
    super(
      `Ukuran gambar setelah kompresi (${sizeMb} MiB) masih melebihi batas ${maxMb} MiB.`,
    );
    this.name = "ImageTooLargeError";
    this.sizeBytes = sizeBytes;
    this.maxBytes = maxBytes;
  }
}

export function computeTargetSize(
  width: number,
  height: number,
  max: number = DEFAULT_MAX_DIMENSION,
): { width: number; height: number } {
  if (width <= 0 || height <= 0 || max <= 0) {
    return { width: Math.max(0, width), height: Math.max(0, height) };
  }
  if (width <= max && height <= max) {
    return { width: Math.round(width), height: Math.round(height) };
  }
  const scale = max / Math.max(width, height);
  return {
    width: Math.max(1, Math.round(width * scale)),
    height: Math.max(1, Math.round(height * scale)),
  };
}

export function validateImageMimeType(mimeType: string): void {
  if (!mimeType.startsWith("image/")) {
    throw new Error("File harus berupa gambar (JPEG, PNG, atau WebP).");
  }
}

export function assertCompressedSize(
  sizeBytes: number,
  maxBytes: number = DEFAULT_MAX_BYTES,
): void {
  if (sizeBytes > maxBytes) {
    throw new ImageTooLargeError(sizeBytes, maxBytes);
  }
}

function canvasToBlob(
  canvas: HTMLCanvasElement,
  type: string,
  quality: number,
): Promise<Blob | null> {
  return new Promise((resolve) => {
    canvas.toBlob((blob) => resolve(blob), type, quality);
  });
}

function replaceExtension(filename: string, ext: string): string {
  const base = filename.replace(/\.[^.]+$/, "");
  return `${base || "dokumen"}.${ext}`;
}

export interface CompressImageOptions {
  maxDimension?: number;
  maxBytes?: number;
  quality?: number;
}

export async function compressImage(
  file: File,
  opts: CompressImageOptions = {},
): Promise<File> {
  validateImageMimeType(file.type);

  const maxDimension = opts.maxDimension ?? DEFAULT_MAX_DIMENSION;
  const maxBytes = opts.maxBytes ?? DEFAULT_MAX_BYTES;
  const quality = opts.quality ?? DEFAULT_IMAGE_QUALITY;

  const bitmap = await createImageBitmap(file);
  try {
    const target = computeTargetSize(bitmap.width, bitmap.height, maxDimension);
    const canvas = document.createElement("canvas");
    canvas.width = target.width;
    canvas.height = target.height;
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      throw new Error("Canvas 2D context tidak tersedia.");
    }
    ctx.drawImage(bitmap, 0, 0, target.width, target.height);

    let blob = await canvasToBlob(canvas, "image/webp", quality);
    if (blob?.type !== "image/webp") {
      blob = await canvasToBlob(canvas, "image/jpeg", quality);
    }
    if (!blob) {
      throw new Error("Gagal mengompresi gambar.");
    }

    assertCompressedSize(blob.size, maxBytes);

    const ext = blob.type === "image/webp" ? "webp" : "jpg";
    return new File([blob], replaceExtension(file.name, ext), {
      type: blob.type,
    });
  } finally {
    bitmap.close();
  }
}
