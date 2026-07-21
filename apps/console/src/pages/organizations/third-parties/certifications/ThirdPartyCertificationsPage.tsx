// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import {
  certifications,
  objectEntries,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import {
  Badge,
  Button,
  Card,
  Combobox,
  ComboboxItem,
  IconCrossLargeX,
  IconPlusLarge,
} from "@probo/ui";
import { clsx } from "clsx";
import { useState } from "react";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { ThirdPartyCertificationsPageQuery } from "#/__generated__/core/ThirdPartyCertificationsPageQuery.graphql";
import { useThirdPartyForm } from "#/hooks/forms/useThirdPartyForm";

export const thirdPartyCertificationsPageQuery = graphql`
  query ThirdPartyCertificationsPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        name
        canUpdate: permission(action: "core:thirdParty:update")
        ...useThirdPartyFormFragment
      }
    }
  }
`;

interface ThirdPartyCertificationsPageProps {
  queryRef: PreloadedQuery<ThirdPartyCertificationsPageQuery>;
}

export default function ThirdPartyCertificationsPage(
  props: ThirdPartyCertificationsPageProps,
) {
  const data = usePreloadedQuery<ThirdPartyCertificationsPageQuery>(thirdPartyCertificationsPageQuery, props.queryRef);
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = data.node;

  const { t } = useTranslation();
  const { control, handleSubmit } = useThirdPartyForm(thirdParty);

  usePageTitle(t("thirdPartyCertificationsPage.pageTitle", { name: thirdParty.name }));

  return (
    <form
      className="space-y-4"
      onSubmit={thirdParty.canUpdate
        ? e => void handleSubmit(e)
        : undefined}
    >
      <Card padded>
        <Controller
          control={control}
          name="certifications"
          render={({ field }) => (
            <Certifications
              onValueChange={field.onChange}
              value={field.value ?? []}
              readOnly={!thirdParty.canUpdate}
            />
          )}
        />
      </Card>
      {thirdParty.canUpdate && (
        <div className="flex justify-end">
          <Button type="submit">{t("thirdPartyCertificationsPage.actions.update")}</Button>
        </div>
      )}
    </form>
  );
}

interface CertificationsProps {
  value: string[];
  onValueChange: (value: string[]) => void;
  readOnly?: boolean;
}

function Certifications(props: CertificationsProps) {
  const categorizedCertifications = Object.values(certifications).flat();
  const { t } = useTranslation();
  const [animateBadge, setAnimateBadge] = useState(false);
  const categories = objectEntries(certifications)
    .map(
      ([key, value]) =>
        [key, value.filter(c => props.value.includes(c))] as const,
    )
    .filter(([, certs]) => certs.length > 0);
  categories.push([
    "custom",
    props.value.filter(c => !categorizedCertifications.includes(c)),
  ]);

  const addCertificate = (name: string) => {
    setAnimateBadge(true);
    props.onValueChange([...props.value, name]);
  };

  const removeCertificate = (name: string) => {
    setAnimateBadge(true);
    props.onValueChange(props.value.filter(v => v !== name));
  };

  return (
    <div className="space-y-6">
      {categories.map(([key, certs]) => (
        <div key={key} className="space-y-2">
          <div className="text-sm font-medium text-txt-secondary">
            {t(`thirdPartyCertificationsPage.categories.${key}`)}
          </div>
          <div className="flex flex-wrap gap-2">
            {certs.map(certification => (
              <Badge asChild size="md" key={certification}>
                {props.readOnly
                  ? (
                      <span>{certification}</span>
                    )
                  : (
                      <button
                        onClick={() => removeCertificate(certification)}
                        type="button"
                        className={clsx(
                          "hover:bg-subtle-hover cursor-pointer",
                          animateBadge
                          && "starting:opacity-0 starting:w-0 w-max transition-all duration-500 starting:bg-accent",
                        )}
                      >
                        {certification}
                        <div className="w-0 overflow-hidden group-hover:w-4 duration-200">
                          <IconCrossLargeX size={12} />
                        </div>
                      </button>
                    )}
              </Badge>
            ))}
          </div>
        </div>
      ))}
      {!props.readOnly && (
        <CertificationInput
          certifications={categorizedCertifications.filter(
            c => !props.value.includes(c),
          )}
          onAdd={addCertificate}
        />
      )}
    </div>
  );
}

function CertificationInput({
  certifications,
  onAdd,
}: {
  certifications: string[];
  onAdd: (name: string) => void;
}) {
  const { t } = useTranslation();
  const [search, setSearch] = useState("");
  const isCustom = !certifications.includes(search.trim());
  const filteredCertifications = certifications.filter(c =>
    c.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div className="flex items-center gap-2">
      <Combobox
        autoSelect
        resetValueOnHide
        onSelect={onAdd}
        onSearch={setSearch}
        placeholder={t("thirdPartyCertificationsPage.placeholders.add")}
      >
        {filteredCertifications.map(certification => (
          <ComboboxItem key={certification} value={certification}>
            {certification}
          </ComboboxItem>
        ))}
        {isCustom && search.trim().length >= 2 && (
          <ComboboxItem value={search.trim()}>
            <IconPlusLarge size={20} />
            {t("thirdPartyCertificationsPage.addCustom", { name: search })}
          </ComboboxItem>
        )}
      </Combobox>
    </div>
  );
}
