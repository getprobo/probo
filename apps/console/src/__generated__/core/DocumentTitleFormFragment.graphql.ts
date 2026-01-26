/**
 * @generated SignedSource<<a02660daf2bb259b6312b0cdf5fbfc17>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentTitleFormFragment$data = {
  readonly canUpdate: boolean;
  readonly id: string;
  readonly title: string;
  readonly " $fragmentType": "DocumentTitleFormFragment";
};
export type DocumentTitleFormFragment$key = {
  readonly " $data"?: DocumentTitleFormFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentTitleFormFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentTitleFormFragment",
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
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document:update\")"
    }
  ],
  "type": "Document",
  "abstractKey": null
};

(node as any).hash = "5fa870524b66e0b68c0f898982e56ce0";

export default node;
