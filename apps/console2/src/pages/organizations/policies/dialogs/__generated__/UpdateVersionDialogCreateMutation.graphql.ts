/**
 * @generated SignedSource<<7b7cbab5e73007ace1e7d07970b633e8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PolicyStatus = "DRAFT" | "PUBLISHED";
export type PolicyVersionSignatureState = "REQUESTED" | "SIGNED";
export type CreateDraftPolicyVersionInput = {
  policyID: string;
};
export type UpdateVersionDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateDraftPolicyVersionInput;
};
export type UpdateVersionDialogCreateMutation$data = {
  readonly createDraftPolicyVersion: {
    readonly policyVersionEdge: {
      readonly node: {
        readonly content: string;
        readonly id: string;
        readonly publishedAt: any | null | undefined;
        readonly signatures: {
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly state: PolicyVersionSignatureState;
            };
          }>;
        };
        readonly status: PolicyStatus;
        readonly updatedAt: any;
        readonly version: number;
      };
    };
  };
};
export type UpdateVersionDialogCreateMutation = {
  response: UpdateVersionDialogCreateMutation$data;
  variables: UpdateVersionDialogCreateMutation$variables;
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
},
v4 = {
  "alias": null,
  "args": null,
  "concreteType": "PolicyVersionEdge",
  "kind": "LinkedField",
  "name": "policyVersionEdge",
  "plural": false,
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
          "name": "content",
          "storageKey": null
        },
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
          "name": "publishedAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "version",
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
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "UpdateVersionDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDraftPolicyVersionPayload",
        "kind": "LinkedField",
        "name": "createDraftPolicyVersion",
        "plural": false,
        "selections": [
          (v4/*: any*/)
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
    "name": "UpdateVersionDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDraftPolicyVersionPayload",
        "kind": "LinkedField",
        "name": "createDraftPolicyVersion",
        "plural": false,
        "selections": [
          (v4/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "policyVersionEdge",
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
    "cacheID": "c5b0f449a70a4043368de55bba176329",
    "id": null,
    "metadata": {},
    "name": "UpdateVersionDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation UpdateVersionDialogCreateMutation(\n  $input: CreateDraftPolicyVersionInput!\n) {\n  createDraftPolicyVersion(input: $input) {\n    policyVersionEdge {\n      node {\n        id\n        content\n        status\n        publishedAt\n        version\n        updatedAt\n        signatures(first: 100) {\n          edges {\n            node {\n              id\n              state\n            }\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6a451392a86776d33c376bfdcea7613c";

export default node;
