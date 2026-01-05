/**
 * @generated SignedSource<<c8b89765dcf362beef95c0f17902a77b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
import { FragmentRefs } from "relay-runtime";
export type DocumentRowFragment$data = {
  readonly classification: DocumentClassification;
  readonly documentType: DocumentType;
  readonly id: string;
  readonly signed: boolean;
  readonly title: string;
  readonly updatedAt: string;
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
      "name": "documentType",
      "storageKey": null
    },
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
      "kind": "ScalarField",
      "name": "signed",
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
  "type": "SignableDocument",
  "abstractKey": null
};

(node as any).hash = "43959adf3e4545cce78ba06f34c82cd2";

export default node;
