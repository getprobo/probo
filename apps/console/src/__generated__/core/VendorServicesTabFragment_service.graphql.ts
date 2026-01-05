/**
 * @generated SignedSource<<f8640f48ff7cc9e6d9431195f6aec66b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorServicesTabFragment_service$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly createdAt: string;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly name: string;
  readonly updatedAt: string;
  readonly " $fragmentType": "VendorServicesTabFragment_service";
};
export type VendorServicesTabFragment_service$key = {
  readonly " $data"?: VendorServicesTabFragment_service$data;
  readonly " $fragmentSpreads": FragmentRefs<"VendorServicesTabFragment_service">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VendorServicesTabFragment_service",
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
      "name": "description",
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
          "value": "core:vendor-service:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:vendor-service:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:vendor-service:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:vendor-service:delete\")"
    }
  ],
  "type": "VendorService",
  "abstractKey": null
};

(node as any).hash = "6371ec5f43c5d090c72fa50ca7ec1686";

export default node;
