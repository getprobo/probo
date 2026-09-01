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

import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { BindingsList_viewer$key } from "#/__generated__/core/BindingsList_viewer.graphql";
import type { BindingsListRefetchQuery } from "#/__generated__/core/BindingsListRefetchQuery.graphql";
import { PageHeader } from "#/pages/_components/PageHeader";

import { BindingListItem } from "./BindingListItem";
import { BindingsEmpty } from "./BindingsEmpty";
import { bindingsList } from "./variants";

const bindingsListViewerFragment = graphql`
  fragment BindingsList_viewer on Viewer
  @refetchable(queryName: "BindingsListRefetchQuery")
  @throwOnFieldError {
    probotIdentityBindings {
      id
      ...BindingListItem_binding
    }
  }
`;

export interface BindingsListProps {
  viewerKey: BindingsList_viewer$key;
}

export function BindingsList({ viewerKey }: BindingsListProps) {
  const { t } = useTranslation("bindings");
  const { t: tApp } = useTranslation();
  const slots = bindingsList();
  const [viewer, refetchBindings] = useRefetchableFragment<
    BindingsListRefetchQuery,
    BindingsList_viewer$key
  >(bindingsListViewerFragment, viewerKey);
  const [, startTransition] = useTransition();
  const bindings = viewer.probotIdentityBindings;

  function handleBindingDeleted() {
    startTransition(() => {
      refetchBindings({}, { fetchPolicy: "store-and-network" });
    });
  }

  return (
    <>
      <PageHeader
        homeLabel={tApp("homePage.breadcrumb")}
        currentLabel={t("list.breadcrumb")}
        title={t("list.title")}
      />
      {bindings.length === 0
        ? <BindingsEmpty />
        : (
            <div className={slots.body()}>
              <div className={slots.frame()}>
                <Table variant="ghost" className={slots.table()}>
                  <TableHeader className={slots.header()}>
                    <TableRow>
                      <TableColumnHeaderCell>
                        {t("list.columns.account")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell>
                        {t("list.columns.action")}
                      </TableColumnHeaderCell>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {bindings.map(binding => (
                      <BindingListItem
                        key={binding.id}
                        bindingKey={binding}
                        onDeleted={handleBindingDeleted}
                      />
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}
    </>
  );
}
