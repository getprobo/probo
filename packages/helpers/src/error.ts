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

export interface GraphQLError {
  message?: string;
  extensions?: {
    code?: string;
  };
  source?: {
    errors?: Array<{ message: string; extensions?: { code?: string } }>;
  };
}

export function formatError(title: string, error: GraphQLError | GraphQLError[]): string {
  const messages: string[] = [];

  if (Array.isArray(error)) {
    messages.push(...error.map((e) => e.message).filter(Boolean) as string[]);
  } else if (error.source?.errors && Array.isArray(error.source.errors)) {
    messages.push(...error.source.errors.map((e) => e.message).filter(Boolean));
  } else if (error.message) {
    messages.push(error.message);
  }

  if (messages.length === 0) {
    return title;
  }

  const errorList = messages.join(", ");

  return `${title}: ${errorList}${errorList.endsWith('.') ? '' : '.'}`;
}
