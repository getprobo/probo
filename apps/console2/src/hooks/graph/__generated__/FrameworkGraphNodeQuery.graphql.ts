/**
 * @generated SignedSource<<9d5d654dbc062e8a2a82adeed4c0839e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworkGraphNodeQuery$variables = {
  frameworkId: string;
};
export type FrameworkGraphNodeQuery$data = {
  readonly node: {
    readonly id?: string;
    readonly name?: string;
    readonly " $fragmentSpreads": FragmentRefs<"FrameworkDetailPageFragment">;
  };
};
export type FrameworkGraphNodeQuery = {
  response: FrameworkGraphNodeQuery$data;
  variables: FrameworkGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "frameworkId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "frameworkId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v6 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "endCursor",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "hasNextPage",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v9 = {
  "kind": "ClientExtension",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "__id",
      "storageKey": null
    }
  ]
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "FrameworkGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "FrameworkDetailPageFragment"
              }
            ],
            "type": "Framework",
            "abstractKey": null
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "FrameworkGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v4/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v5/*: any*/),
              {
                "alias": null,
                "args": (v6/*: any*/),
                "concreteType": "ControlConnection",
                "kind": "LinkedField",
                "name": "controls",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ControlEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Control",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "referenceId",
                            "storageKey": null
                          },
                          (v3/*: any*/),
                          (v5/*: any*/),
                          {
                            "alias": null,
                            "args": (v6/*: any*/),
                            "concreteType": "MeasureConnection",
                            "kind": "LinkedField",
                            "name": "measures",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "MeasureEdge",
                                "kind": "LinkedField",
                                "name": "edges",
                                "plural": true,
                                "selections": [
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "Measure",
                                    "kind": "LinkedField",
                                    "name": "node",
                                    "plural": false,
                                    "selections": [
                                      (v2/*: any*/),
                                      (v3/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "kind": "ScalarField",
                                        "name": "state",
                                        "storageKey": null
                                      },
                                      (v4/*: any*/)
                                    ],
                                    "storageKey": null
                                  },
                                  (v7/*: any*/)
                                ],
                                "storageKey": null
                              },
                              (v8/*: any*/),
                              (v9/*: any*/)
                            ],
                            "storageKey": "measures(first:100)"
                          },
                          {
                            "alias": null,
                            "args": (v6/*: any*/),
                            "filters": null,
                            "handle": "connection",
                            "key": "FrameworkDetailPage_measures",
                            "kind": "LinkedHandle",
                            "name": "measures"
                          },
                          {
                            "alias": null,
                            "args": (v6/*: any*/),
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
                                      (v2/*: any*/),
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
                                        "name": "createdAt",
                                        "storageKey": null
                                      },
                                      {
                                        "alias": null,
                                        "args": null,
                                        "kind": "ScalarField",
                                        "name": "documentType",
                                        "storageKey": null
                                      },
                                      {
                                        "alias": null,
                                        "args": [
                                          {
                                            "kind": "Literal",
                                            "name": "first",
                                            "value": 1
                                          }
                                        ],
                                        "concreteType": "DocumentVersionConnection",
                                        "kind": "LinkedField",
                                        "name": "versions",
                                        "plural": false,
                                        "selections": [
                                          {
                                            "alias": null,
                                            "args": null,
                                            "concreteType": "DocumentVersionEdge",
                                            "kind": "LinkedField",
                                            "name": "edges",
                                            "plural": true,
                                            "selections": [
                                              {
                                                "alias": null,
                                                "args": null,
                                                "concreteType": "DocumentVersion",
                                                "kind": "LinkedField",
                                                "name": "node",
                                                "plural": false,
                                                "selections": [
                                                  (v2/*: any*/),
                                                  {
                                                    "alias": null,
                                                    "args": null,
                                                    "kind": "ScalarField",
                                                    "name": "status",
                                                    "storageKey": null
                                                  }
                                                ],
                                                "storageKey": null
                                              }
                                            ],
                                            "storageKey": null
                                          }
                                        ],
                                        "storageKey": "versions(first:1)"
                                      },
                                      (v4/*: any*/)
                                    ],
                                    "storageKey": null
                                  },
                                  (v7/*: any*/)
                                ],
                                "storageKey": null
                              },
                              (v8/*: any*/),
                              (v9/*: any*/)
                            ],
                            "storageKey": "documents(first:100)"
                          },
                          {
                            "alias": null,
                            "args": (v6/*: any*/),
                            "filters": null,
                            "handle": "connection",
                            "key": "FrameworkDetailPage_documents",
                            "kind": "LinkedHandle",
                            "name": "documents"
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "controls(first:100)"
              }
            ],
            "type": "Framework",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "9af171f23ae160b6725af13fdebaa07a",
    "id": null,
    "metadata": {},
    "name": "FrameworkGraphNodeQuery",
    "operationKind": "query",
    "text": "query FrameworkGraphNodeQuery(\n  $frameworkId: ID!\n) {\n  node(id: $frameworkId) {\n    __typename\n    ... on Framework {\n      id\n      name\n      ...FrameworkDetailPageFragment\n    }\n    id\n  }\n}\n\nfragment FrameworkDetailPageFragment on Framework {\n  id\n  name\n  description\n  controls(first: 100) {\n    edges {\n      node {\n        id\n        referenceId\n        name\n        description\n        measures(first: 100) {\n          edges {\n            node {\n              id\n              ...LinkedMeasuresCardFragment\n              __typename\n            }\n            cursor\n          }\n          pageInfo {\n            endCursor\n            hasNextPage\n          }\n        }\n        documents(first: 100) {\n          edges {\n            node {\n              id\n              ...LinkedDocumentsCardFragment\n              __typename\n            }\n            cursor\n          }\n          pageInfo {\n            endCursor\n            hasNextPage\n          }\n        }\n      }\n    }\n  }\n}\n\nfragment LinkedDocumentsCardFragment on Document {\n  id\n  title\n  createdAt\n  documentType\n  versions(first: 1) {\n    edges {\n      node {\n        id\n        status\n      }\n    }\n  }\n}\n\nfragment LinkedMeasuresCardFragment on Measure {\n  id\n  name\n  state\n}\n"
  }
};
})();

(node as any).hash = "69a32dc8fbcaefbbaeb02dd2b89752c4";

export default node;
