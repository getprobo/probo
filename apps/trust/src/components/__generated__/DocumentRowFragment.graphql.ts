/**
 * @generated SignedSource<<396f9d87342c909ca373f60a1e6f7ff3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentRowFragment$data = {
  readonly hasUserRequestedAccess: boolean;
  readonly id: string;
  readonly isUserAuthorized: boolean;
  readonly title: string;
  readonly " $fragmentType": "DocumentRowFragment";
};
export type DocumentRowFragment$key = {
  readonly " $data"?: DocumentRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentRowFragment",
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
      "name": "title",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "isUserAuthorized",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "hasUserRequestedAccess",
      "storageKey": null
    }
  ],
  "type": "Document",
  "abstractKey": null
};

(node as any).hash = "1b0f0b9cae621268d72b0fd0646aa672";

export default node;
