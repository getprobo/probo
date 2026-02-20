/**
 * @generated SignedSource<<97bc9b075b82bb1543c80843430ced5d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ElectronicSignatureStatus = "ACCEPTED" | "COMPLETED" | "FAILED" | "PENDING" | "PROCESSING";
export type TrustCenterAccessState = "ACTIVE" | "INACTIVE";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageAccessListItemFragment$data = {
  readonly activeCount: number;
  readonly canUpdate: boolean;
  readonly createdAt: string;
  readonly email: string;
  readonly id: string;
  readonly name: string;
  readonly ndaSignature: {
    readonly status: ElectronicSignatureStatus;
  } | null | undefined;
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
      "concreteType": "ElectronicSignature",
      "kind": "LinkedField",
      "name": "ndaSignature",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "status",
          "storageKey": null
        }
      ],
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

(node as any).hash = "9047d3d3f232d43c6f951c74b0395b58";

export default node;
