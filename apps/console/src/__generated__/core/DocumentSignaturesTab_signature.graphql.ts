/**
 * @generated SignedSource<<02f9338b7a451ae446c87d09ea71a71d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturesTab_signature$data = {
  readonly canCancel: boolean;
  readonly id: string;
  readonly requestedAt: any;
  readonly signedAt: any | null | undefined;
  readonly signedBy: {
    readonly fullName: string;
    readonly primaryEmailAddress: any;
  };
  readonly state: DocumentVersionSignatureState;
  readonly " $fragmentType": "DocumentSignaturesTab_signature";
};
export type DocumentSignaturesTab_signature$key = {
  readonly " $data"?: DocumentSignaturesTab_signature$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesTab_signature">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignaturesTab_signature",
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
    },
    {
      "alias": "canCancel",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document-version:request-signature"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document-version:request-signature\")"
    }
  ],
  "type": "DocumentVersionSignature",
  "abstractKey": null
};

(node as any).hash = "bbfdd78015e97032862da41aad6beeae";

export default node;
