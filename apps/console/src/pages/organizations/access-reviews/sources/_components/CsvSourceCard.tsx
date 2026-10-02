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

import { FileCsvIcon } from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { TonedCard } from "#/components/TonedCard/TonedCard";

interface CsvSourceCardProps {
  organizationId: string;
}

export function CsvSourceCard({ organizationId }: CsvSourceCardProps) {
  const { t } = useTranslation();

  return (
    <TonedCard
      tone="sand"
      size={2}
      icon={<FileCsvIcon />}
      lead={(
        <Heading level={3} size={3} weight="medium" highContrast>
          {t("addAccessReviewSourceDialog.csv.title")}
        </Heading>
      )}
      control={(
        <ButtonLink
          to={`/organizations/${organizationId}/access-reviews/sources/new/csv`}
          variant="solid"
          size={1}
        >
          {t("addAccessReviewSourceDialog.actions.open")}
        </ButtonLink>
      )}
    >
      <Text size={2} color="faint">
        {t("addAccessReviewSourceDialog.csv.description")}
      </Text>
    </TonedCard>
  );
}
