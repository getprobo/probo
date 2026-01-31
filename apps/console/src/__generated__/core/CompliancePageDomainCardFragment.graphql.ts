/**
 * @generated SignedSource<<0e36707d7a08e91112f9f44e787b4844>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageDomainCardFragment$data = {
  readonly canDelete: boolean;
  readonly domain: string;
  readonly sslStatus: SSLStatus;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDomainDialogFragment">;
  readonly " $fragmentType": "CompliancePageDomainCardFragment";
};
export type CompliancePageDomainCardFragment$key = {
  readonly " $data"?: CompliancePageDomainCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDomainCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageDomainCardFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "domain",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "sslStatus",
      "storageKey": null
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:custom-domain:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:custom-domain:delete\")"
    },
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "CompliancePageDomainDialogFragment"
    }
  ],
  "type": "CustomDomain",
  "abstractKey": null
};

(node as any).hash = "db079748369f6103bb263de4e8921c76";

export default node;
