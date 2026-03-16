/**
 * @generated SignedSource<<df6b24e4cabfa9f10e1fa41aaaf42b68>>
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
    readonly id: string;
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
  "name": "id",
  "storageKey": null
},
v1 = {
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
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileName",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileUrl",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "NDAPageQuery",
    "selections": [
      (v1/*: any*/),
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
                (v2/*: any*/)
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
                (v3/*: any*/),
                (v4/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "ElectronicSignature",
                  "kind": "LinkedField",
                  "name": "viewerSignature",
                  "plural": false,
                  "selections": [
                    (v5/*: any*/)
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
      (v1/*: any*/),
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
              (v2/*: any*/),
              (v0/*: any*/)
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
              (v3/*: any*/),
              (v4/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "ElectronicSignature",
                "kind": "LinkedField",
                "name": "viewerSignature",
                "plural": false,
                "selections": [
                  (v5/*: any*/),
                  (v0/*: any*/),
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
          (v0/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "3fd076a45e4a2eae4a56fc34aeafa815",
    "id": null,
    "metadata": {},
    "name": "NDAPageQuery",
    "operationKind": "query",
    "text": "query NDAPageQuery {\n  viewer {\n    id\n  }\n  currentTrustCenter {\n    organization {\n      name\n      id\n    }\n    nonDisclosureAgreement {\n      fileName\n      fileUrl\n      viewerSignature {\n        status\n        id\n      }\n    }\n    ...NDAPageFragment\n    id\n  }\n}\n\nfragment NDAPageFragment on TrustCenter {\n  nonDisclosureAgreement {\n    viewerSignature {\n      id\n      status\n      consentText\n      lastError\n    }\n  }\n  id\n}\n"
  }
};
})();

(node as any).hash = "048b0469d90d78f35ce7ed1bfe7edbde";

export default node;
