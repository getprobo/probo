/**
 * @generated SignedSource<<266bdec238fe7397f3cff10dc7da8fc6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentRowFragment$data = {
  readonly id: string;
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
    }
  ],
  "type": "Document",
  "abstractKey": null
};

(node as any).hash = "437d68812bcc68724d2e347fea22b5e3";

export default node;
