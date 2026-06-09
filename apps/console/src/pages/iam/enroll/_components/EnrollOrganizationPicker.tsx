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

interface OrganizationOption {
  id: string;
  name: string;
}

interface EnrollOrganizationPickerProps {
  organizations: OrganizationOption[];
  selectedOrganizationId: string | null;
  onChange: (organizationID: string) => void;
}

export function EnrollOrganizationPicker(
  {
    organizations,
    selectedOrganizationId,
    onChange,
  }: EnrollOrganizationPickerProps,
) {
  const { __ } = useTranslate();

  return (
    <section className="space-y-4">
      <label htmlFor="organization-select" className="space-y-2 text-xs font-medium text-txt-secondary">
        <span>{__("Organization")}</span>
        <select
          id="organization-select"
          value={selectedOrganizationId ?? ""}
          onChange={event => onChange(event.target.value)}
          className="w-full rounded-lg border border-border-low bg-level-1 px-3 py-2 text-sm text-txt-primary outline-none transition-colors focus:border-border-solid"
        >
          {organizations.map(organization => (
            <option key={organization.id} value={organization.id}>
              {organization.name}
            </option>
          ))}
        </select>
      </label>
    </section>
  );
}
