/**
 * @generated SignedSource<<5a43f6367b3b03e57c58a1ab616e867b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturePlaceholder_personFragment$data = {
  readonly fullName: string;
  readonly id: string;
  readonly primaryEmailAddress: string;
  readonly " $fragmentType": "DocumentSignaturePlaceholder_personFragment";
};
export type DocumentSignaturePlaceholder_personFragment$key = {
  readonly " $data"?: DocumentSignaturePlaceholder_personFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturePlaceholder_personFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignaturePlaceholder_personFragment",
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
      "name": "fullName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "primaryEmailAddress",
      "storageKey": null
    }
  ],
  "type": "People",
  "abstractKey": null
};

(node as any).hash = "4b77b54fc11c37a0e1c50f4d482b0f18";

export default node;
