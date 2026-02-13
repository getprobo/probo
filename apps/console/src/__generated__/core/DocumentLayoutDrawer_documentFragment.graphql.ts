/**
 * @generated SignedSource<<547f6c8756b51ff46a73158262f78e8e>>
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
  readonly approvers: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly fullName: string;
        readonly id: string;
      };
    }>;
  };
  readonly canUpdate: boolean;
  readonly documentType: DocumentType;
  readonly id: string;
  readonly " $fragmentType": "DocumentLayoutDrawer_documentFragment";
};
export type DocumentLayoutDrawer_documentFragment$key = {
  readonly " $data"?: DocumentLayoutDrawer_documentFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentLayoutDrawer_documentFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentLayoutDrawer_documentFragment",
  "selections": [
    (v0/*: any*/),
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
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 100
        }
      ],
      "concreteType": "ProfileConnection",
      "kind": "LinkedField",
      "name": "approvers",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ProfileEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Profile",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
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
          "storageKey": null
        }
      ],
      "storageKey": "approvers(first:100)"
    }
  ],
  "type": "Document",
  "abstractKey": null
};
})();

(node as any).hash = "ffc32318ed0df6b2992eca91bef297e5";

export default node;
