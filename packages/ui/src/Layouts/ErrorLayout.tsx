// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { PropsWithChildren } from "react";

import { IconCircleInfo } from "../Atoms/Icons";

type Props = PropsWithChildren<{
  title?: string;
  description?: string;
}>;

export function ErrorLayout({ title, description, children }: Props) {
  return (
    <div className="min-h-screen w-full gap-2 flex flex-col items-center justify-center">
      <IconCircleInfo className="text-txt-danger" size={40} />
      <h1 className="text-4xl font-bold">{title}</h1>
      <p className="text-txt-secondary">{description}</p>
      {children}
    </div>
  );
}
