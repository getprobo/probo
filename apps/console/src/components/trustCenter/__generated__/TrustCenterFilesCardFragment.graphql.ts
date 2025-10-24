/**
 * @generated SignedSource<<0defaaf1ce3544420e8fbd9c9f3af139>>
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
  readonly category: string;
  readonly createdAt: any;
  readonly fileUrl: string;
  readonly id: string;
  readonly name: string;
  readonly trustCenterVisibility: TrustCenterVisibility;
  readonly updatedAt: any;
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
    }
  ],
  "type": "TrustCenterFile",
  "abstractKey": null
};

(node as any).hash = "33cb01782ca37fc776cd8e5dfde20f76";

export default node;
