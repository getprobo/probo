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
import { useForm } from "react-hook-form";

interface FormValues {
  displayName: string;
  description: string;
}

interface TrackerResourceRowEditProps {
  displayName: string;
  description: string;
  isUpdating: boolean;
  onSave: (data: { displayName: string; description: string }) => void;
  onCancel: () => void;
}

export function TrackerResourceRowEdit({
  displayName,
  description,
  isUpdating,
  onSave,
  onCancel,
}: TrackerResourceRowEditProps) {
  const { __ } = useTranslate();

  const { register, handleSubmit } = useForm<FormValues>({
    defaultValues: {
      displayName,
      description,
    },
  });

  const onSubmit = (data: FormValues) => {
    onSave({
      displayName: data.displayName,
      description: data.description,
    });
  };

  return (
    <Tr>
      <Td />
      <Td className="pr-3">
        <Input
          {...register("displayName")}
          placeholder={__("Display name")}
        />
      </Td>
      <Td className="pr-3" colSpan={2}>
        <div className="flex items-center gap-2">
          <Input
            {...register("description")}
            placeholder={__("Description")}
            className="flex-1"
          />
          <Button
            onClick={() => void handleSubmit(onSubmit)()}
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
