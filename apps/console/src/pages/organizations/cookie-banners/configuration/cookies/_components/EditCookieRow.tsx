// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Button, Input, Td, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EditCookieRowFragment$key } from "#/__generated__/core/EditCookieRowFragment.graphql";

import type { CookieEntry } from "./CategorySection";

export const editCookieRowFragment = graphql`
  fragment EditCookieRowFragment on Cookie {
    name
    duration
    description
  }
`;

interface EditCookieRowProps {
  cookieKey: EditCookieRowFragment$key;
  isUpdating: boolean;
  onSave: (cookie: CookieEntry) => void;
  onCancel: () => void;
}

export function EditCookieRow({
  cookieKey,
  isUpdating,
  onSave,
  onCancel,
}: EditCookieRowProps) {
  const { __ } = useTranslate();
  const cookie = useFragment(editCookieRowFragment, cookieKey);
  const [form, setForm] = useState<CookieEntry>({
    name: cookie.name,
    duration: cookie.duration,
    description: cookie.description,
  });

  return (
    <Tr>
      <Td className="pr-3">
        <Input
          value={form.name}
          onChange={e => setForm({ ...form, name: e.target.value })}
          placeholder={__("Cookie name")}
        />
      </Td>
      <Td className="pr-3">
        <Input
          value={form.duration}
          onChange={e => setForm({ ...form, duration: e.target.value })}
          placeholder={__("e.g. 1 year")}
        />
      </Td>
      <Td className="pr-3">
        <Input
          value={form.description}
          onChange={e => setForm({ ...form, description: e.target.value })}
          placeholder={__("Description")}
        />
      </Td>
      <Td>
        <div className="flex items-center gap-1">
          <Button
            onClick={() => onSave(form)}
            disabled={isUpdating}
          >
            {__("Save")}
          </Button>
          <Button
            variant="secondary"
            onClick={onCancel}
          >
            {__("Cancel")}
          </Button>
        </div>
      </Td>
    </Tr>
  );
}
