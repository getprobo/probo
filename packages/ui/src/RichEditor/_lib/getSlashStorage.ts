// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import type { Editor } from "@tiptap/react";

import type { SlashCommandStorage } from "../SlashCommandExtension";

export function getSlashStorage(
  editor: Editor,
): SlashCommandStorage | undefined {
  return (editor.storage as unknown as Record<string, unknown>).slashCommand as
    | SlashCommandStorage
    | undefined;
}
