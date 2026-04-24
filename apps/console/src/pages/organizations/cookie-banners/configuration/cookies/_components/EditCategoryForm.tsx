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
import { Button, Input, Textarea } from "@probo/ui";
import { useState } from "react";

const GCM_CONSENT_TYPES = [
  "analytics_storage",
  "ad_storage",
  "ad_user_data",
  "ad_personalization",
  "functionality_storage",
  "personalization_storage",
  "security_storage",
] as const;

interface EditCategoryFormProps {
  name: string;
  slug: string;
  description: string;
  gcmConsentTypes: string[];
  posthogConsent: boolean;
  isUpdating: boolean;
  onSave: (name: string, slug: string, description: string, gcmConsentTypes: string[], posthogConsent: boolean) => void;
  onCancel: () => void;
}

export function EditCategoryForm({
  name,
  slug,
  description,
  gcmConsentTypes,
  posthogConsent,
  isUpdating,
  onSave,
  onCancel,
}: EditCategoryFormProps) {
  const { __ } = useTranslate();
  const [editName, setEditName] = useState(name);
  const [editSlug, setEditSlug] = useState(slug);
  const [editDescription, setEditDescription] = useState(description);
  const [editGcmTypes, setEditGcmTypes] = useState<string[]>(gcmConsentTypes);
  const [editPosthogConsent, setEditPosthogConsent] = useState(posthogConsent);

  const toggleGcmType = (type: string) => {
    setEditGcmTypes(prev =>
      prev.includes(type)
        ? prev.filter(t => t !== type)
        : [...prev, type],
    );
  };

  return (
    <div className="space-y-3">
      <Input
        value={editName}
        onChange={e => setEditName(e.target.value)}
        placeholder={__("Category name")}
      />
      <Input
        value={editSlug}
        onChange={e => setEditSlug(e.target.value)}
        placeholder={__("Category slug")}
        pattern="[a-z0-9]+(-[a-z0-9]+)*"
      />
      <Textarea
        value={editDescription}
        onChange={e => setEditDescription(e.target.value)}
        placeholder={__("Category description")}
        rows={2}
      />
      <div>
        <label className="text-sm font-medium">
          {__("Google Consent Mode")}
        </label>
        <p className="text-xs text-muted-foreground mb-2">
          {__("Select the Google Consent Mode signals this category controls.")}
        </p>
        <div className="flex flex-wrap gap-2">
          {GCM_CONSENT_TYPES.map(type => (
            <label
              key={type}
              className="flex items-center gap-1.5 text-xs cursor-pointer"
            >
              <input
                type="checkbox"
                checked={editGcmTypes.includes(type)}
                onChange={() => toggleGcmType(type)}
                className="rounded"
              />
              <code className="font-mono">{type}</code>
            </label>
          ))}
        </div>
      </div>
      <div>
        <label className="text-sm font-medium">
          {__("PostHog")}
        </label>
        <p className="text-xs text-muted-foreground mb-2">
          {__("Control PostHog tracking consent based on this category.")}
        </p>
        <label className="flex items-center gap-1.5 text-xs cursor-pointer">
          <input
            type="checkbox"
            checked={editPosthogConsent}
            onChange={() => setEditPosthogConsent(prev => !prev)}
            className="rounded"
          />
          <span>{__("Opt in/out of PostHog tracking")}</span>
        </label>
      </div>
      <div className="flex items-center gap-2">
        <Button
          onClick={() => {
            if (!/^[a-z0-9]+(-[a-z0-9]+)*$/.test(editSlug)) return;
            onSave(editName, editSlug, editDescription, editGcmTypes, editPosthogConsent);
          }}
          disabled={isUpdating}
        >
          {isUpdating ? __("Saving...") : __("Save")}
        </Button>
        <Button
          variant="secondary"
          onClick={onCancel}
        >
          {__("Cancel")}
        </Button>
      </div>
    </div>
  );
}
