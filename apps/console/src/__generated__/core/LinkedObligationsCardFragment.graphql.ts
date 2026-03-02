/**
 * @generated SignedSource<<f5cbc176c2905c2eed828f946fc581c1>>
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
  readonly requirement: string | null | undefined;
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
      "name": "requirement",
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

(node as any).hash = "67ae0e6b0df090bb5a708bf7e029f4c5";

export default node;
