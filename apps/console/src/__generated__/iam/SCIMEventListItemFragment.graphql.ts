/**
 * @generated SignedSource<<6bc4cdb83f5825b5d606fe579fa8f7a8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SCIMEventListItemFragment$data = {
  readonly createdAt: string;
  readonly errorMessage: string | null | undefined;
  readonly ipAddress: string;
  readonly method: string;
  readonly path: string;
  readonly profile: {
    readonly fullName: string;
  } | null | undefined;
  readonly statusCode: number;
  readonly " $fragmentType": "SCIMEventListItemFragment";
};
export type SCIMEventListItemFragment$key = {
  readonly " $data"?: SCIMEventListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SCIMEventListItemFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SCIMEventListItemFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "method",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "path",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "statusCode",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "errorMessage",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "ipAddress",
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
      "concreteType": "Profile",
      "kind": "LinkedField",
      "name": "profile",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fullName",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "SCIMEvent",
  "abstractKey": null
};

(node as any).hash = "b38af9ac9b660d3aab75e00e2c55dfb6";

export default node;
