/**
 * @generated SignedSource<<ab516d36b92795ed8d94d69153957f93>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
import { FragmentRefs } from "relay-runtime";
export type DomainCardFragment$data = {
  readonly canDelete: boolean;
  readonly domain: string;
  readonly sslStatus: SSLStatus;
  readonly " $fragmentSpreads": FragmentRefs<"DomainDialogFragment">;
  readonly " $fragmentType": "DomainCardFragment";
};
export type DomainCardFragment$key = {
  readonly " $data"?: DomainCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DomainCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DomainCardFragment",
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
      "name": "DomainDialogFragment"
    }
  ],
  "type": "CustomDomain",
  "abstractKey": null
};

(node as any).hash = "91cc7871d5ac257f2540568a9a10ce0f";

export default node;
