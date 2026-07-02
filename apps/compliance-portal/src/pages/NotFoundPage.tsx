// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { Link } from "@probo/ui/src/v2/Button/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

// Catch-all page for portal paths that match no route, so an unknown URL renders
// an explicit not-found state inside the layout instead of an empty body.
export default function NotFoundPage() {
  const { t } = useTranslation();

  return (
    <HeaderBand>
      <div className="flex flex-col items-start gap-4">
        <Heading level={1} size={7} weight="medium" highContrast>
          {t("notFound.title")}
        </Heading>
        <Text size={2} color="neutral">
          {t("notFound.description")}
        </Text>
        <Link to="/" variant="soft" color="neutral" highContrast size={2}>
          {t("notFound.backHome")}
        </Link>
      </div>
    </HeaderBand>
  );
}
