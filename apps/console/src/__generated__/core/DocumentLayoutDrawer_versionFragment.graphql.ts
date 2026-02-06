/**
 * @generated SignedSource<<02a8fd2b19d7f7fdc180b373fa0b17c9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type DocumentStatus = "DRAFT" | "PUBLISHED";
import { FragmentRefs } from "relay-runtime";
export type DocumentLayoutDrawer_versionFragment$data = {
  readonly classification: DocumentClassification;
  readonly id: string;
  readonly owner: {
    readonly fullName: string;
    readonly id: string;
  };
  readonly publishedAt: string | null | undefined;
  readonly status: DocumentStatus;
  readonly updatedAt: string;
  readonly version: number;
  readonly " $fragmentType": "DocumentLayoutDrawer_versionFragment";
};
export type DocumentLayoutDrawer_versionFragment$key = {
  readonly " $data"?: DocumentLayoutDrawer_versionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentLayoutDrawer_versionFragment">;
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
  "name": "DocumentLayoutDrawer_versionFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "classification",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Profile",
      "kind": "LinkedField",
      "name": "owner",
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
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "version",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "status",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "updatedAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "publishedAt",
      "storageKey": null
    }
  ],
  "type": "DocumentVersion",
  "abstractKey": null
};
})();

(node as any).hash = "f5e9183b163586bab18475a19843b186";

export default node;
