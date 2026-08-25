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

import { dateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { visitorPage } from "../variants";

export function CompliancePortalVisitorProfileCard({
  fullName,
  emailAddress,
  createdAt,
}: {
  fullName: string;
  emailAddress: string;
  createdAt: string;
}) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const { profile, person, identity, joined } = visitorPage();

  return (
    <Card variant="ghost" size={2} padding="none" className={profile()}>
      <div className={person()}>
        <Avatar
          size={5}
          variant="soft"
          color="gold"
          fallback={fullName.charAt(0).toUpperCase() || "?"}
        />
        <div className={identity()}>
          <Heading level={2} size={4} weight="medium" highContrast className="truncate">
            {fullName}
          </Heading>
          <Text size={2} color="gold" className="truncate">
            {emailAddress}
          </Text>
        </div>
      </div>
      <Separator />
      <div className={joined()}>
        <Text size={1} color="faint">
          {t("visitorPage.joinedOn", { date: dateFormat(i18n.language, createdAt) })}
        </Text>
      </div>
    </Card>
  );
}
