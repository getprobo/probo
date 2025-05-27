/**
 * @generated SignedSource<<0471409a1735c49fc0b05b5e63e97853>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreatePolicyInput = {
  content: string;
  organizationId: string;
  ownerId: string;
  title: string;
};
export type CreatePolicyDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreatePolicyInput;
};
export type CreatePolicyDialogMutation$data = {
  readonly createPolicy: {
    readonly policyEdge: {
      readonly node: {
        readonly " $fragmentSpreads": FragmentRefs<"PoliciesPageRowFragment">;
      };
    };
  };
};
export type CreatePolicyDialogMutation = {
  response: CreatePolicyDialogMutation$data;
  variables: CreatePolicyDialogMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "CreatePolicyDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreatePolicyPayload",
        "kind": "LinkedField",
        "name": "createPolicy",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "PolicyEdge",
            "kind": "LinkedField",
            "name": "policyEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Policy",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "PoliciesPageRowFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "CreatePolicyDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreatePolicyPayload",
        "kind": "LinkedField",
        "name": "createPolicy",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "PolicyEdge",
            "kind": "LinkedField",
            "name": "policyEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Policy",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
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
                    "name": "description",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "updatedAt",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "People",
                    "kind": "LinkedField",
                    "name": "owner",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "fullName",
                        "storageKey": null
                      }
                    ],
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
                    "concreteType": "PolicyVersionConnection",
                    "kind": "LinkedField",
                    "name": "versions",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "PolicyVersionEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "PolicyVersion",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "status",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "first",
                                    "value": 100
                                  }
                                ],
                                "concreteType": "PolicyVersionSignatureConnection",
                                "kind": "LinkedField",
                                "name": "signatures",
                                "plural": false,
                                "selections": [
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "PolicyVersionSignatureEdge",
                                    "kind": "LinkedField",
                                    "name": "edges",
                                    "plural": true,
                                    "selections": [
                                      {
                                        "alias": null,
                                        "args": null,
                                        "concreteType": "PolicyVersionSignature",
                                        "kind": "LinkedField",
                                        "name": "node",
                                        "plural": false,
                                        "selections": [
                                          (v3/*: any*/),
                                          {
                                            "alias": null,
                                            "args": null,
                                            "kind": "ScalarField",
                                            "name": "state",
                                            "storageKey": null
                                          }
                                        ],
                                        "storageKey": null
                                      }
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": "signatures(first:100)"
                              }
                            ],
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": "versions(first:1)"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "policyEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "7ca1b2811e06da7ba97064950fb8f632",
    "id": null,
    "metadata": {},
    "name": "CreatePolicyDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreatePolicyDialogMutation(\n  $input: CreatePolicyInput!\n) {\n  createPolicy(input: $input) {\n    policyEdge {\n      node {\n        ...PoliciesPageRowFragment\n        id\n      }\n    }\n  }\n}\n\nfragment PoliciesPageRowFragment on Policy {\n  id\n  title\n  description\n  updatedAt\n  owner {\n    id\n    fullName\n  }\n  versions(first: 1) {\n    edges {\n      node {\n        id\n        status\n        signatures(first: 100) {\n          edges {\n            node {\n              id\n              state\n            }\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6c749af4def68b742bc5eb80dd120c8a";

export default node;
