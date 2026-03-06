/**
 * @generated SignedSource<<a006942c31a97ab7a7ac843a12295b27>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TrustGraphCurrentNewsQuery$variables = Record<PropertyKey, never>;
export type TrustGraphCurrentNewsQuery$data = {
  readonly currentTrustCenter: {
    readonly id: string;
    readonly updates: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly body: string;
          readonly id: string;
          readonly title: string;
          readonly updatedAt: any;
        };
      }>;
    };
  } | null | undefined;
};
export type TrustGraphCurrentNewsQuery = {
  response: TrustGraphCurrentNewsQuery$data;
  variables: TrustGraphCurrentNewsQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "TrustCenter",
    "kind": "LinkedField",
    "name": "currentTrustCenter",
    "plural": false,
    "selections": [
      (v0/*: any*/),
      {
        "alias": null,
        "args": [
          {
            "kind": "Literal",
            "name": "first",
            "value": 50
          }
        ],
        "concreteType": "MailingListUpdateConnection",
        "kind": "LinkedField",
        "name": "updates",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "MailingListUpdateEdge",
            "kind": "LinkedField",
            "name": "edges",
            "plural": true,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "MailingListUpdate",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v0/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "title",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "body",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "updatedAt",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": "updates(first:50)"
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustGraphCurrentNewsQuery",
    "selections": (v1/*: any*/),
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "TrustGraphCurrentNewsQuery",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "33f19ffe9740fe50be600b97a4768d00",
    "id": null,
    "metadata": {},
    "name": "TrustGraphCurrentNewsQuery",
    "operationKind": "query",
    "text": "query TrustGraphCurrentNewsQuery {\n  currentTrustCenter {\n    id\n    updates(first: 50) {\n      edges {\n        node {\n          id\n          title\n          body\n          updatedAt\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "daedd5964c39110b4e29a3c63a2a3014";

export default node;
