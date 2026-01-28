/**
 * @generated SignedSource<<e798de547d8804086770073da5d19f76>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageFileListItem_fileFragment$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly category: string;
  readonly createdAt: string;
  readonly fileUrl: string;
  readonly id: string;
  readonly name: string;
  readonly trustCenterVisibility: TrustCenterVisibility;
  readonly " $fragmentType": "CompliancePageFileListItem_fileFragment";
};
export type CompliancePageFileListItem_fileFragment$key = {
  readonly " $data"?: CompliancePageFileListItem_fileFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageFileListItem_fileFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageFileListItem_fileFragment",
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
      "name": "category",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fileUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "trustCenterVisibility",
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
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center-file:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-file:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center-file:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-file:delete\")"
    }
  ],
  "type": "TrustCenterFile",
  "abstractKey": null
};

(node as any).hash = "0e5caf3d8c62ca0cd3ed8b37787e2b86";

export default node;
