// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

function isTableSeparatorCell(cell: string): boolean {
  const trimmed = cell.trim();
  if (trimmed.length < 3) return false;

  let start = 0;
  let end = trimmed.length;
  if (trimmed[0] === ":") start = 1;
  if (trimmed[end - 1] === ":") end -= 1;
  if (end - start < 3) return false;

  for (let i = start; i < end; i++) {
    if (trimmed[i] !== "-") return false;
  }
  return true;
}

function isTableSeparatorLine(line: string): boolean {
  const trimmed = line.trim();
  if (trimmed.length < 3) return false;

  let s = trimmed;
  if (s[0] === "|") s = s.slice(1);
  if (s[s.length - 1] === "|") s = s.slice(0, -1);
  if (s.trim().length === 0) return false;

  const cells = s.split("|");
  for (const cell of cells) {
    if (!isTableSeparatorCell(cell)) return false;
  }
  return true;
}

function isWordChar(c: number): boolean {
  return (
    (c >= 48 && c <= 57) ||
    (c >= 65 && c <= 90) ||
    (c >= 97 && c <= 122) ||
    c === 95
  );
}

function hasMarkdownLine(line: string): boolean {
  if (line.length === 0) return false;
  const ch = line[0];

  if (ch === "`" && line.startsWith("```")) {
    for (let i = 3; i < line.length; i++) {
      if (!isWordChar(line.charCodeAt(i))) return false;
    }
    return true;
  }

  if (ch === "#") {
    let i = 1;
    while (i < line.length && i < 6 && line[i] === "#") i++;
    return i <= 6 && i < line.length && (line[i] === " " || line[i] === "\t");
  }

  if (ch === ">" && line.length > 1 && (line[1] === " " || line[1] === "\t")) {
    return true;
  }

  if (
    (ch === "-" || ch === "+" || ch === "*") &&
    line.length > 1 &&
    (line[1] === " " || line[1] === "\t")
  ) {
    return true;
  }

  if (ch >= "0" && ch <= "9") {
    let i = 1;
    while (i < line.length && line[i] >= "0" && line[i] <= "9") i++;
    if (
      i < line.length &&
      line[i] === "." &&
      i + 1 < line.length &&
      (line[i + 1] === " " || line[i + 1] === "\t")
    ) {
      return true;
    }
  }

  if (
    line.length >= 3 &&
    (ch === "-" || ch === "_" || ch === "*") &&
    line[1] === ch &&
    line[2] === ch
  ) {
    for (let i = 3; i < line.length; i++) {
      if (line[i] !== " " && line[i] !== "\t") return false;
    }
    return true;
  }

  return false;
}

export function hasMarkdown(text: string): boolean {
  const lines = text.split("\n");
  for (const line of lines) {
    if (hasMarkdownLine(line) || isTableSeparatorLine(line)) return true;
  }
  return false;
}
