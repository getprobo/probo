/**
 * @generated SignedSource<<5d977fa59296dafac2297640e7dd3293>>
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
    readonly complianceNews: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly body: string;
          readonly id: string;
          readonly title: string;
          readonly updatedAt: any;
        };
      }>;
    };
    readonly id: string;
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
        "concreteType": "ComplianceNewsConnection",
        "kind": "LinkedField",
        "name": "complianceNews",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ComplianceNewsEdge",
            "kind": "LinkedField",
            "name": "edges",
            "plural": true,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ComplianceNews",
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
        "storageKey": "complianceNews(first:50)"
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
    "cacheID": "1d714eb435fde8fa61313d497438def1",
    "id": null,
    "metadata": {},
    "name": "TrustGraphCurrentNewsQuery",
    "operationKind": "query",
    "text": "query TrustGraphCurrentNewsQuery {\n  currentTrustCenter {\n    id\n    complianceNews(first: 50) {\n      edges {\n        node {\n          id\n          title\n          body\n          updatedAt\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "0a8de5e7274f52a6b47508b98b3b445b";

export default node;
