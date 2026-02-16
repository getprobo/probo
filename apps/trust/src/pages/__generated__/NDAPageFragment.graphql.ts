/**
 * @generated SignedSource<<93f8676826dfa28c1681da68a6ac1cf2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ElectronicSignatureStatus = "ACCEPTED" | "COMPLETED" | "FAILED" | "PENDING" | "PROCESSING";
import { FragmentRefs } from "relay-runtime";
export type NDAPageFragment$data = {
  readonly id: string;
  readonly ndaSignature: {
    readonly consentText: string;
    readonly id: string;
    readonly lastError: string | null | undefined;
    readonly status: ElectronicSignatureStatus;
  };
  readonly " $fragmentType": "NDAPageFragment";
};
export type NDAPageFragment$key = {
  readonly " $data"?: NDAPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"NDAPageFragment">;
};

import NDAPageRefetchQuery_graphql from './NDAPageRefetchQuery.graphql';

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
  "metadata": {
    "refetch": {
      "connection": null,
      "fragmentPathInResult": [
        "node"
      ],
      "operation": NDAPageRefetchQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "NDAPageFragment",
  "selections": [
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "ElectronicSignature",
        "kind": "LinkedField",
        "name": "ndaSignature",
        "plural": false,
        "selections": [
          (v0/*: any*/),
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
            "name": "consentText",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "lastError",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    },
    (v0/*: any*/)
  ],
  "type": "TrustCenter",
  "abstractKey": null
};
})();

(node as any).hash = "b53a5f455508ed45124b7d09cfb573da";

export default node;
