/**
 * @generated SignedSource<<10a0ee1dfa19239c5c5820f3d58a1e19>>
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
    readonly nonDisclosureAgreement: {
      readonly fileName: string;
      readonly fileUrl: string;
      readonly viewerSignature: {
        readonly status: ElectronicSignatureStatus;
      } | null | undefined;
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
  "name": "fileName",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileUrl",
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
            {
              "alias": null,
              "args": null,
              "concreteType": "NonDisclosureAgreement",
              "kind": "LinkedField",
              "name": "nonDisclosureAgreement",
              "plural": false,
              "selections": [
                (v2/*: any*/),
                (v3/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "ElectronicSignature",
                  "kind": "LinkedField",
                  "name": "viewerSignature",
                  "plural": false,
                  "selections": [
                    (v4/*: any*/)
                  ],
                  "storageKey": null
                }
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
          {
            "alias": null,
            "args": null,
            "concreteType": "NonDisclosureAgreement",
            "kind": "LinkedField",
            "name": "nonDisclosureAgreement",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "ElectronicSignature",
                "kind": "LinkedField",
                "name": "viewerSignature",
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
    "cacheID": "e1ab3a5206a4ea200ebdb991023a0e99",
    "id": null,
    "metadata": {},
    "name": "NDAPageQuery",
    "operationKind": "query",
    "text": "query NDAPageQuery {\n  viewer {\n    fullName\n    id\n  }\n  currentTrustCenter {\n    organization {\n      name\n      id\n    }\n    nonDisclosureAgreement {\n      fileName\n      fileUrl\n      viewerSignature {\n        status\n        id\n      }\n    }\n    ...NDAPageFragment\n    id\n  }\n}\n\nfragment NDAPageFragment on TrustCenter {\n  nonDisclosureAgreement {\n    viewerSignature {\n      id\n      status\n      consentText\n      lastError\n    }\n  }\n  id\n}\n"
  }
};
})();

(node as any).hash = "4f0f0dd672862fb6c3fb5ef2eff80cd9";

export default node;
