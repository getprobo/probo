/**
 * @generated SignedSource<<f347bcc313be984e835dd6518091a71c>>
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
  readonly statusCode: number;
  readonly userName: string;
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
      "kind": "ScalarField",
      "name": "userName",
      "storageKey": null
    }
  ],
  "type": "SCIMEvent",
  "abstractKey": null
};

(node as any).hash = "9df2e986643ab90589ebc6e4631cde24";

export default node;
