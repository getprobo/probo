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
