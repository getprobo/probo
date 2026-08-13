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

import { usePageTitle } from "@probo/hooks";
import { Card, Slack } from "@probo/ui";
import { useTransition } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
  useRefetchableFragment,
} from "react-relay";

import type { EmployeeBindingsPage_viewer$key } from "#/__generated__/core/EmployeeBindingsPage_viewer.graphql";
import type { EmployeeBindingsPageQuery } from "#/__generated__/core/EmployeeBindingsPageQuery.graphql";
import type { EmployeeBindingsPageRefetchQuery } from "#/__generated__/core/EmployeeBindingsPageRefetchQuery.graphql";

import { EmployeeBindingListItem } from "./_components/EmployeeBindingListItem";

const employeeBindingsPageViewerFragment = graphql`
  fragment EmployeeBindingsPage_viewer on Viewer
  @refetchable(queryName: "EmployeeBindingsPageRefetchQuery") {
    probotIdentityBindings {
      id
      ...EmployeeBindingListItem_binding
    }
  }
`;

export const employeeBindingsPageQuery = graphql`
  query EmployeeBindingsPageQuery {
    viewer @required(action: THROW) {
      ...EmployeeBindingsPage_viewer
    }
  }
`;

interface EmployeeBindingsPageProps {
  queryRef: PreloadedQuery<EmployeeBindingsPageQuery>;
}

export function EmployeeBindingsPage({ queryRef }: EmployeeBindingsPageProps) {
  const { t } = useTranslation();

  usePageTitle(t("employeeBindingsPage.title"));

  const data = usePreloadedQuery<EmployeeBindingsPageQuery>(
    employeeBindingsPageQuery,
    queryRef,
  );

  const [, startTransition] = useTransition();

  const [viewerData, refetchBindings] = useRefetchableFragment<
    EmployeeBindingsPageRefetchQuery,
    EmployeeBindingsPage_viewer$key
  >(employeeBindingsPageViewerFragment, data.viewer);

  const binding = viewerData.probotIdentityBindings[0];

  const handleBindingDeleted = () => {
    startTransition(() => {
      refetchBindings({}, { fetchPolicy: "store-and-network" });
    });
  };

  if (!binding) {
    return (
      <Card padded>
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 flex items-center justify-center bg-subtle rounded">
            <Slack className="h-6 w-6" />
          </div>
          <div>
            <h3 className="text-base font-semibold">
              {t("employeeBindingsPage.empty.title")}
            </h3>
            <p className="text-sm text-txt-tertiary">
              {t("employeeBindingsPage.empty.description")}
            </p>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <EmployeeBindingListItem
      bindingKey={binding}
      onDeleted={handleBindingDeleted}
    />
  );
}
