/**
 * @generated SignedSource<<d6d18b9131581530d367e10380f14077>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type TrustCenterAccessState = "ACTIVE" | "INACTIVE";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageAccessListItemFragment$data = {
  readonly activeCount: number;
  readonly canUpdate: boolean;
  readonly createdAt: string;
  readonly email: string;
  readonly hasAcceptedNonDisclosureAgreement: boolean;
  readonly id: string;
  readonly name: string;
  readonly pendingRequestCount: number;
  readonly state: TrustCenterAccessState;
  readonly " $fragmentType": "CompliancePageAccessListItemFragment";
};
export type CompliancePageAccessListItemFragment$key = {
  readonly " $data"?: CompliancePageAccessListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAccessListItemFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageAccessListItemFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "email",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "createdAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "state",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "activeCount",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "pendingRequestCount",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "hasAcceptedNonDisclosureAgreement",
      "storageKey": null
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center-access:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-access:update\")"
    }
  ],
  "type": "TrustCenterAccess",
  "abstractKey": null
};

(node as any).hash = "2492ed4967d57febe8f1153f75280410";

export default node;
