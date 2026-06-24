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
