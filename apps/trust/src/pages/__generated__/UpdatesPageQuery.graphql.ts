/**
 * @generated SignedSource<<4560049874109fe98528d64249673d23>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdatesPageQuery$variables = Record<PropertyKey, never>;
export type UpdatesPageQuery$data = {
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
export type UpdatesPageQuery = {
  response: UpdatesPageQuery$data;
  variables: UpdatesPageQuery$variables;
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
    "name": "UpdatesPageQuery",
    "selections": (v1/*: any*/),
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "UpdatesPageQuery",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "5d323cbbdf9bfec294bb22dd6a6be735",
    "id": null,
    "metadata": {},
    "name": "UpdatesPageQuery",
    "operationKind": "query",
    "text": "query UpdatesPageQuery {\n  currentTrustCenter {\n    id\n    updates(first: 50) {\n      edges {\n        node {\n          id\n          title\n          body\n          updatedAt\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5c0e074ce562f24bd841cc750f492988";

export default node;
