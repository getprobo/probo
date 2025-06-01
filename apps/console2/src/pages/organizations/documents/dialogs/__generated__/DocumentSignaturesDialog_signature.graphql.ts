/**
 * @generated SignedSource<<71e44f5802c3eaba88dde399af20ed89>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturesDialog_signature$data = {
  readonly id: string;
  readonly requestedAt: any;
  readonly signedAt: any | null | undefined;
  readonly signedBy: {
    readonly fullName: string;
    readonly primaryEmailAddress: string;
  };
  readonly state: DocumentVersionSignatureState;
  readonly " $fragmentType": "DocumentSignaturesDialog_signature";
};
export type DocumentSignaturesDialog_signature$key = {
  readonly " $data"?: DocumentSignaturesDialog_signature$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesDialog_signature">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignaturesDialog_signature",
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
      "name": "state",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "signedAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "requestedAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "signedBy",
      "plural": false,
      "selections": [
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
      "storageKey": null
    }
  ],
  "type": "DocumentVersionSignature",
  "abstractKey": null
};

(node as any).hash = "0db843d4ebd21f44a3d126e85da8cf26";

export default node;
