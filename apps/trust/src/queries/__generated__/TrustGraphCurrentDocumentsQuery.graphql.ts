/**
 * @generated SignedSource<<dabe16493c003a98ed8fd555efc85ddc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustGraphCurrentDocumentsQuery$variables = Record<PropertyKey, never>;
export type TrustGraphCurrentDocumentsQuery$data = {
  readonly currentTrustCenter: {
    readonly documents: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly documentType: DocumentType;
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"DocumentRowFragment">;
        };
      }>;
    };
    readonly id: string;
    readonly organization: {
      readonly name: string;
    };
    readonly trustCenterFiles: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly category: string;
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"TrustCenterFileRowFragment">;
        };
      }>;
    };
  } | null | undefined;
};
export type TrustGraphCurrentDocumentsQuery = {
  response: TrustGraphCurrentDocumentsQuery$data;
  variables: TrustGraphCurrentDocumentsQuery$variables;
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
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v2 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "documentType",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "category",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "isUserAuthorized",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasUserRequestedAccess",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustGraphCurrentDocumentsQuery",
    "selections": [
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
            "args": (v2/*: any*/),
            "concreteType": "DocumentConnection",
            "kind": "LinkedField",
            "name": "documents",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "DocumentEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Document",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      (v3/*: any*/),
                      {
                        "args": null,
                        "kind": "FragmentSpread",
                        "name": "DocumentRowFragment"
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "documents(first:50)"
          },
          {
            "alias": null,
            "args": (v2/*: any*/),
            "concreteType": "TrustCenterFileConnection",
            "kind": "LinkedField",
            "name": "trustCenterFiles",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterFileEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterFile",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      (v4/*: any*/),
                      {
                        "args": null,
                        "kind": "FragmentSpread",
                        "name": "TrustCenterFileRowFragment"
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "trustCenterFiles(first:50)"
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "TrustGraphCurrentDocumentsQuery",
    "selections": [
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
            "args": null,
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              (v0/*: any*/)
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": (v2/*: any*/),
            "concreteType": "DocumentConnection",
            "kind": "LinkedField",
            "name": "documents",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "DocumentEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Document",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "title",
                        "storageKey": null
                      },
                      (v5/*: any*/),
                      (v6/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "documents(first:50)"
          },
          {
            "alias": null,
            "args": (v2/*: any*/),
            "concreteType": "TrustCenterFileConnection",
            "kind": "LinkedField",
            "name": "trustCenterFiles",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterFileEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterFile",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      (v4/*: any*/),
                      (v1/*: any*/),
                      (v5/*: any*/),
                      (v6/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "trustCenterFiles(first:50)"
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "c9c4cb18da98c40646fbbe21c2c5b46f",
    "id": null,
    "metadata": {},
    "name": "TrustGraphCurrentDocumentsQuery",
    "operationKind": "query",
    "text": "query TrustGraphCurrentDocumentsQuery {\n  currentTrustCenter {\n    id\n    organization {\n      name\n      id\n    }\n    documents(first: 50) {\n      edges {\n        node {\n          id\n          documentType\n          ...DocumentRowFragment\n        }\n      }\n    }\n    trustCenterFiles(first: 50) {\n      edges {\n        node {\n          id\n          category\n          ...TrustCenterFileRowFragment\n        }\n      }\n    }\n  }\n}\n\nfragment DocumentRowFragment on Document {\n  id\n  title\n  isUserAuthorized\n  hasUserRequestedAccess\n}\n\nfragment TrustCenterFileRowFragment on TrustCenterFile {\n  id\n  name\n  isUserAuthorized\n  hasUserRequestedAccess\n}\n"
  }
};
})();

(node as any).hash = "0900835ad8e59a40a2907fb787b936ef";

export default node;
