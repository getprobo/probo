// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
      <Td className="pr-3" colSpan={4}>
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
