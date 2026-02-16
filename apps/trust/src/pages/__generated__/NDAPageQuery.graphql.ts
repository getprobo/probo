/**
 * @generated SignedSource<<f1cae8462501cdf634eb02eb143c3806>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ElectronicSignatureStatus = "ACCEPTED" | "COMPLETED" | "FAILED" | "PENDING" | "PROCESSING";
export type NDAPageQuery$variables = Record<PropertyKey, never>;
export type NDAPageQuery$data = {
  readonly currentTrustCenter: {
    readonly ndaFileName: string | null | undefined;
    readonly ndaFileUrl: string | null | undefined;
    readonly ndaSignature: {
      readonly status: ElectronicSignatureStatus;
    } | null | undefined;
    readonly organization: {
      readonly name: string;
    };
    readonly " $fragmentSpreads": FragmentRefs<"NDAPageFragment">;
  };
  readonly viewer: {
    readonly fullName: string;
  } | null | undefined;
};
export type NDAPageQuery = {
  response: NDAPageQuery$data;
  variables: NDAPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "ndaFileUrl",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "ndaFileName",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "NDAPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v0/*: any*/)
        ],
        "storageKey": null
      },
      {
        "kind": "RequiredField",
        "field": {
          "alias": null,
          "args": null,
          "concreteType": "TrustCenter",
          "kind": "LinkedField",
          "name": "currentTrustCenter",
          "plural": false,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Organization",
              "kind": "LinkedField",
              "name": "organization",
              "plural": false,
              "selections": [
                (v1/*: any*/)
              ],
              "storageKey": null
            },
            (v2/*: any*/),
            (v3/*: any*/),
            {
              "alias": null,
              "args": null,
              "concreteType": "ElectronicSignature",
              "kind": "LinkedField",
              "name": "ndaSignature",
              "plural": false,
              "selections": [
                (v4/*: any*/)
              ],
              "storageKey": null
            },
            {
              "args": null,
              "kind": "FragmentSpread",
              "name": "NDAPageFragment"
            }
          ],
          "storageKey": null
        },
        "action": "THROW"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "NDAPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          (v5/*: any*/)
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "currentTrustCenter",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              (v5/*: any*/)
            ],
            "storageKey": null
          },
          (v2/*: any*/),
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "ElectronicSignature",
            "kind": "LinkedField",
            "name": "ndaSignature",
            "plural": false,
            "selections": [
              (v4/*: any*/),
              (v5/*: any*/),
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
          (v5/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "c0b175c435c5730e390fcfde596d8bc3",
    "id": null,
    "metadata": {},
    "name": "NDAPageQuery",
    "operationKind": "query",
    "text": "query NDAPageQuery {\n  viewer {\n    fullName\n    id\n  }\n  currentTrustCenter {\n    organization {\n      name\n      id\n    }\n    ndaFileUrl\n    ndaFileName\n    ndaSignature {\n      status\n      id\n    }\n    ...NDAPageFragment\n    id\n  }\n}\n\nfragment NDAPageFragment on TrustCenter {\n  ndaSignature {\n    id\n    status\n    consentText\n    lastError\n  }\n  id\n}\n"
  }
};
})();

(node as any).hash = "a78bf8b65feba4627ecb265ca882f645";

export default node;
