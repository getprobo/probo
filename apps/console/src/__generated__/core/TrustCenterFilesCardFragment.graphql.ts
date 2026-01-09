/**
 * @generated SignedSource<<f9b0d1c5f1852bef8797dcf34920e519>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
import { FragmentRefs } from "relay-runtime";
export type TrustCenterFilesCardFragment$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly category: string;
  readonly createdAt: string;
  readonly fileUrl: string;
  readonly id: string;
  readonly name: string;
  readonly trustCenterVisibility: TrustCenterVisibility;
  readonly updatedAt: string;
  readonly " $fragmentType": "TrustCenterFilesCardFragment";
};
export type TrustCenterFilesCardFragment$key = {
  readonly " $data"?: TrustCenterFilesCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterFilesCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "TrustCenterFilesCardFragment",
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
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "updatedAt",
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

(node as any).hash = "7b9e482868fc5d815923fbefdbafc8bb";

export default node;
