/**
 * @generated SignedSource<<99908411212decf464eafa45b1fc0e04>>
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
  readonly nonDisclosureAgreement: {
    readonly viewerSignature: {
      readonly consentText: string;
      readonly id: string;
      readonly lastError: string | null | undefined;
      readonly status: ElectronicSignatureStatus;
    };
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
        "concreteType": "NonDisclosureAgreement",
        "kind": "LinkedField",
        "name": "nonDisclosureAgreement",
        "plural": false,
        "selections": [
          {
            "kind": "RequiredField",
            "field": {
              "alias": null,
              "args": null,
              "concreteType": "ElectronicSignature",
              "kind": "LinkedField",
              "name": "viewerSignature",
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

(node as any).hash = "503859b8c00b27214f9f3253bffcb670";

export default node;
