/**
 * @generated SignedSource<<6e7ee5d7dad966afab1e9ddbf60e14f3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorServicesTabFragment_service$data = {
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
    }
  ],
  "type": "VendorService",
  "abstractKey": null
};

(node as any).hash = "276b5545b9e5eb7f9d9f56c4ea2ed352";

export default node;
