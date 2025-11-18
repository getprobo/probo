/**
 * @generated SignedSource<<f3d939f80d769a19b8f8ef64207ae911>>
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
export type EmployeeDocumentsPageRowFragment$data = {
  readonly classification: DocumentClassification;
  readonly documentType: DocumentType;
  readonly id: string;
  readonly signed: boolean;
  readonly title: string;
  readonly updatedAt: any;
  readonly " $fragmentType": "EmployeeDocumentsPageRowFragment";
};
export type EmployeeDocumentsPageRowFragment$key = {
  readonly " $data"?: EmployeeDocumentsPageRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"EmployeeDocumentsPageRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "EmployeeDocumentsPageRowFragment",
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

(node as any).hash = "929301e6f6216fb0678b32061b70dd17";

export default node;
