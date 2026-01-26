/**
 * @generated SignedSource<<927b958ffdaa974f5a5054be4638b60d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
import { FragmentRefs } from "relay-runtime";
export type DocumentLayoutDrawerFragment$data = {
  readonly canUpdate: boolean;
  readonly documentType: DocumentType;
  readonly id: string;
  readonly " $fragmentType": "DocumentLayoutDrawerFragment";
};
export type DocumentLayoutDrawerFragment$key = {
  readonly " $data"?: DocumentLayoutDrawerFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentLayoutDrawerFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentLayoutDrawerFragment",
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
      "name": "documentType",
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

(node as any).hash = "a8f55efdb2682fb35a7bc68f3ef1808a";

export default node;
