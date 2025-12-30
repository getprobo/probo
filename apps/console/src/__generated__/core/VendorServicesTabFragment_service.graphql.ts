/**
 * @generated SignedSource<<2f6dc5bf85b9ad81fe6c9af7955e4217>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorServicesTabFragment_service$data = {
  readonly createdAt: any;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly name: string;
  readonly updatedAt: any;
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
