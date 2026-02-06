/**
 * @generated SignedSource<<c4186b06367f737c6cb7977780eeaaad>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
import { FragmentRefs } from "relay-runtime";
export type DocumentSignatureListItemFragment$data = {
  readonly canCancel: boolean;
  readonly id: string;
  readonly requestedAt: string;
  readonly signedAt: string | null | undefined;
  readonly signedBy: {
    readonly fullName: string;
  };
  readonly state: DocumentVersionSignatureState;
  readonly " $fragmentType": "DocumentSignatureListItemFragment";
};
export type DocumentSignatureListItemFragment$key = {
  readonly " $data"?: DocumentSignatureListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignatureListItemFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignatureListItemFragment",
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
      "concreteType": "Profile",
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
        }
      ],
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
      "alias": "canCancel",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document-version-signature:cancel"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document-version-signature:cancel\")"
    }
  ],
  "type": "DocumentVersionSignature",
  "abstractKey": null
};

(node as any).hash = "59bdf4434147c77415ed636a9b78e154";

export default node;
