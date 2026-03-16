/**
 * @generated SignedSource<<9cd6e4f71ffa86acc1195f307912451d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ObligationStatus = "COMPLIANT" | "NON_COMPLIANT" | "PARTIALLY_COMPLIANT";
import { FragmentRefs } from "relay-runtime";
export type LinkedObligationsCardFragment$data = {
  readonly area: string | null | undefined;
  readonly id: string;
  readonly owner: {
    readonly fullName: string;
  };
  readonly source: string | null | undefined;
  readonly status: ObligationStatus;
  readonly " $fragmentType": "LinkedObligationsCardFragment";
};
export type LinkedObligationsCardFragment$key = {
  readonly " $data"?: LinkedObligationsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"LinkedObligationsCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "LinkedObligationsCardFragment",
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
      "name": "area",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "source",
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
      "concreteType": "Profile",
      "kind": "LinkedField",
      "name": "owner",
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
    }
  ],
  "type": "Obligation",
  "abstractKey": null
};

(node as any).hash = "48a0dbc9d3794cc3d8e7f108d6748c94";

export default node;
