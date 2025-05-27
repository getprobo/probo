/**
 * @generated SignedSource<<4bda7ac839740680ea31b4c20299d9d4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type PolicyVersionSignatureState = "REQUESTED" | "SIGNED";
import { FragmentRefs } from "relay-runtime";
export type PolicySignaturesDialog_signature$data = {
  readonly id: string;
  readonly requestedAt: any;
  readonly signedAt: any | null | undefined;
  readonly signedBy: {
    readonly fullName: string;
    readonly primaryEmailAddress: string;
  };
  readonly state: PolicyVersionSignatureState;
  readonly " $fragmentType": "PolicySignaturesDialog_signature";
};
export type PolicySignaturesDialog_signature$key = {
  readonly " $data"?: PolicySignaturesDialog_signature$data;
  readonly " $fragmentSpreads": FragmentRefs<"PolicySignaturesDialog_signature">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "PolicySignaturesDialog_signature",
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
  "type": "PolicyVersionSignature",
  "abstractKey": null
};

(node as any).hash = "b2e435dc6db3a0df223af98251adcb30";

export default node;
