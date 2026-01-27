/**
 * @generated SignedSource<<f22c42dfa1b8b70ebed9f29c56d42069>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
import { FragmentRefs } from "relay-runtime";
export type DocumentLayoutDrawer_documentFragment$data = {
  readonly canUpdate: boolean;
  readonly documentType: DocumentType;
  readonly id: string;
  readonly " $fragmentType": "DocumentLayoutDrawer_documentFragment";
};
export type DocumentLayoutDrawer_documentFragment$key = {
  readonly " $data"?: DocumentLayoutDrawer_documentFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentLayoutDrawer_documentFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentLayoutDrawer_documentFragment",
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

(node as any).hash = "0f2e539cf30c4e116fe81392605b1481";

export default node;
