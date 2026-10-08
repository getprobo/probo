// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

export const maxAttachmentSize = 50 * 1024 * 1024;

const mimeTypesByExtension: Record<string, readonly string[]> = {
  ".pdf": ["application/pdf"],
  ".doc": ["application/msword"],
  ".docx": ["application/vnd.openxmlformats-officedocument.wordprocessingml.document"],
  ".odt": ["application/vnd.oasis.opendocument.text"],
  ".xls": ["application/vnd.ms-excel"],
  ".xlsx": ["application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"],
  ".ods": ["application/vnd.oasis.opendocument.spreadsheet"],
  ".ppt": ["application/vnd.ms-powerpoint"],
  ".pptx": ["application/vnd.openxmlformats-officedocument.presentationml.presentation"],
  ".odp": ["application/vnd.oasis.opendocument.presentation"],
  ".md": ["text/markdown"],
  ".txt": ["text/plain"],
  ".log": ["text/plain", "text/x-log"],
  ".uri": ["text/uri-list", "text/uri-list; charset=utf-8"],
  ".jpg": ["image/jpeg"],
  ".jpeg": ["image/jpeg"],
  ".png": ["image/png"],
  ".svg": ["image/svg+xml"],
  ".webp": ["image/webp"],
  ".yaml": ["application/yaml", "text/yaml"],
  ".yml": ["application/yaml", "text/yaml"],
  ".json": ["application/json", "text/json"],
  ".csv": ["text/csv", "application/csv"],
};

const imageExtensions = new Set([".jpg", ".jpeg", ".png", ".svg", ".webp"]);

export const imageAccept = ".jpg,.jpeg,.png,.svg,.webp,image/jpeg,image/png,image/svg+xml,image/webp";

export const fileAccept = Object.keys(mimeTypesByExtension)
  .filter(extension => !imageExtensions.has(extension))
  .join(",");

export type AttachmentKind = "image" | "file";

const attachmentPreviews = new Map<string, string>();

export function rememberAttachmentPreview(fileId: string, file: Blob) {
  const current = attachmentPreviews.get(fileId);
  if (current) {
    return current;
  }

  const url = URL.createObjectURL(file);
  attachmentPreviews.set(fileId, url);

  return url;
}

export function attachmentPreviewURL(fileId: string) {
  return attachmentPreviews.get(fileId);
}

export function attachmentPath(fileId: string) {
  return `/api/files/v1/attachments/${encodeURIComponent(fileId)}`;
}

const extensionByMimeType: Record<string, string> = {
  "image/jpeg": ".jpg",
  "image/png": ".png",
  "image/svg+xml": ".svg",
  "image/webp": ".webp",
  "application/pdf": ".pdf",
  "text/plain": ".txt",
  "text/csv": ".csv",
  "application/json": ".json",
};

export function prepareAttachment(file: File) {
  const extension = fileExtension(file.name);
  const mimeType = file.type.toLowerCase();
  const nextExtension = extension || extensionByMimeType[mimeType] || "";
  const nextMimeType = mimeType || mimeTypesByExtension[nextExtension]?.[0] || "";
  if (!nextExtension || !nextMimeType || (extension === nextExtension && mimeType === nextMimeType)) {
    return file;
  }

  const name = extension ? file.name : `pasted${nextExtension}`;

  return new File([file], name, { type: nextMimeType });
}

export function attachmentKind(file: File): AttachmentKind | null {
  if (file.size <= 0 || file.size > maxAttachmentSize) {
    return null;
  }

  const extension = fileExtension(file.name);
  const allowed = mimeTypesByExtension[extension];
  if (!allowed) {
    return null;
  }

  const mimeType = file.type.toLowerCase();
  if (mimeType !== "" && !allowed.includes(mimeType)) {
    return null;
  }

  return imageExtensions.has(extension) ? "image" : "file";
}

export function formatFileSize(size: number) {
  if (size < 1024) {
    return `${size} B`;
  }

  if (size < 1024 * 1024) {
    return `${Math.round(size / 1024)} KB`;
  }

  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function fileExtension(name: string) {
  const dot = name.lastIndexOf(".");
  if (dot < 0) {
    return "";
  }

  return name.slice(dot).toLowerCase();
}
