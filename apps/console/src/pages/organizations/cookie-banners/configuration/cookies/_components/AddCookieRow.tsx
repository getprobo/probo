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

import type { CookieEntry } from "./CategorySection";
import { DurationInput, toMaxAgeSeconds } from "./DurationInput";

interface AddCookieRowProps {
  isUpdating: boolean;
  onSave: (cookie: CookieEntry) => void;
  onCancel: () => void;
}

export function AddCookieRow({
  isUpdating,
  onSave,
  onCancel,
}: AddCookieRowProps) {
  const { __ } = useTranslate();
  const [name, setName] = useState("");
  const [durationValue, setDurationValue] = useState("");
  const [durationUnit, setDurationUnit] = useState("days");
  const [description, setDescription] = useState("");

  const handleSave = () => {
    onSave({
      name,
      maxAgeSeconds: toMaxAgeSeconds(durationValue, durationUnit),
      description,
    });
  };

  return (
    <Tr>
      <Td className="pr-3">
        <Input
          value={name}
          onChange={e => setName(e.target.value)}
          placeholder={__("Cookie name")}
        />
      </Td>
      <Td className="pr-3">
        <DurationInput
          value={durationValue}
          unit={durationUnit}
          onValueChange={setDurationValue}
          onUnitChange={setDurationUnit}
        />
      </Td>
      <Td className="pr-3">
        <Input
          value={description}
          onChange={e => setDescription(e.target.value)}
          placeholder={__("Description")}
        />
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          <Button
            onClick={handleSave}
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
