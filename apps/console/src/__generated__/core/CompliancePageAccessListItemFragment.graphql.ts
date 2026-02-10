/**
 * @generated SignedSource<<e5a5b11ba24baf03ee3f94fa65cf3591>>
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
  readonly canDelete: boolean;
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
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center-access:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-access:delete\")"
    }
  ],
  "type": "TrustCenterAccess",
  "abstractKey": null
};

(node as any).hash = "4e78486a2c8edc2b5f73fb64dbeb4006";

export default node;
