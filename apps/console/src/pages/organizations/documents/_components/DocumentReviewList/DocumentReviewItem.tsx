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

import { relativeDateFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { CardButton } from "@probo/ui/src/v2/Card/CardButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentReviewItem_author$key } from "#/__generated__/core/DocumentReviewItem_author.graphql";

import { documentReviewItem } from "../../variants";

const documentReviewItemAuthorFragment = graphql`
  fragment DocumentReviewItem_author on Profile {
    fullName
    emailAddress
    avatar {
      downloadUrl
    }
  }
`;

interface DocumentReviewItemProps {
  authorKey: DocumentReviewItem_author$key;
  createdAt: string;
  label: string;
  active: boolean;
  children: ReactNode;
  onActivate: () => void;
}

export function DocumentReviewItem({
  authorKey,
  createdAt,
  label,
  active,
  children,
  onActivate,
}: DocumentReviewItemProps) {
  const author = useFragment(documentReviewItemAuthorFragment, authorKey);
  const { i18n, t } = useTranslation("organizations/documents");
  const createdLabel = relativeDateFormat(i18n.language, createdAt)
    || t("details.review.justNow");
  const { root, row, body, header, name, time } = documentReviewItem({ active });

  return (
    <Card variant="soft" padding={1} interactive className={root()}>
      <CardButton aria-label={label} aria-pressed={active} onClick={onActivate} />
      <div className={row()}>
        <Avatar
          name={author.fullName}
          email={author.emailAddress}
          src={author.avatar?.downloadUrl}
          size={1}
          radius="full"
        />
        <div className={body()}>
          <div className={header()}>
            <Text size={1} weight="medium" className={name()}>
              {author.fullName}
            </Text>
            <Text size={1} color="faint" className={time()}>
              {createdLabel}
            </Text>
          </div>
          {children}
        </div>
      </div>
    </Card>
  );
}
