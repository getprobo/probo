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

import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ConnectorListFragment$key } from "#/__generated__/iam/ConnectorListFragment.graphql";

import { GoogleWorkspaceConnector } from "./GoogleWorkspaceConnector";
import { Microsoft365Connector } from "./Microsoft365Connector";

const connectorListFragment = graphql`
  fragment ConnectorListFragment on Organization {
    scimBridgeTypes {
      type
      oauth2Scopes
    }
    scimConfiguration {
      bridge {
        type
      }
      ...GoogleWorkspaceConnectorFragment
      ...Microsoft365ConnectorFragment
    }
  }
`;

export function ConnectorList(props: { fKey: ConnectorListFragment$key }) {
  const { fKey } = props;
  const data = useFragment<ConnectorListFragment$key>(connectorListFragment, fKey);
  const { t } = useTranslation();

  const googleWorkspaceScopes
    = data.scimBridgeTypes.find(info => info.type === "GOOGLE_WORKSPACE")?.oauth2Scopes ?? [];
  const microsoft365Scopes
    = data.scimBridgeTypes.find(info => info.type === "MICROSOFT_365")?.oauth2Scopes ?? [];

  const bridgeType = data.scimConfiguration?.bridge?.type ?? null;
  const showGoogleWorkspace = bridgeType === null || bridgeType === "GOOGLE_WORKSPACE";
  const showMicrosoft365 = bridgeType === null || bridgeType === "MICROSOFT_365";

  return (
    <div className="space-y-4">
      <h2 className="text-base font-medium">{t("connectorList.title")}</h2>
      <p className="text-sm text-txt-secondary">
        {t("connectorList.description")}
      </p>
      {showGoogleWorkspace && (
        <GoogleWorkspaceConnector
          fKey={data.scimConfiguration ?? null}
          oauth2Scopes={googleWorkspaceScopes}
        />
      )}
      {showMicrosoft365 && (
        <Microsoft365Connector
          fKey={data.scimConfiguration ?? null}
          oauth2Scopes={microsoft365Scopes}
        />
      )}
    </div>
  );
}
