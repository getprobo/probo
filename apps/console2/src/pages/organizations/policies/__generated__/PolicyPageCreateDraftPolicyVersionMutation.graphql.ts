/**
 * @generated SignedSource<<0c65ab6f2531f65b9687e3ca13074774>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PolicyStatus = "DRAFT" | "PUBLISHED" | "%future added value";
export type CreateDraftPolicyVersionInput = {
  policyID: string;
};
export type PolicyPageCreateDraftPolicyVersionMutation$variables = {
  input: CreateDraftPolicyVersionInput;
};
export type PolicyPageCreateDraftPolicyVersionMutation$data = {
  readonly createDraftPolicyVersion: {
    readonly policyVersionEdge: {
      readonly node: {
        readonly id: string;
        readonly status: PolicyStatus;
        readonly version: number;
      };
    };
  };
};
export type PolicyPageCreateDraftPolicyVersionMutation = {
  response: PolicyPageCreateDraftPolicyVersionMutation$data;
  variables: PolicyPageCreateDraftPolicyVersionMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "CreateDraftPolicyVersionPayload",
    "kind": "LinkedField",
    "name": "createDraftPolicyVersion",
    "plural": false,
    "selections": [
      {
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
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "id",
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
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "PolicyPageCreateDraftPolicyVersionMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PolicyPageCreateDraftPolicyVersionMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "197f53d67c5393a8163be2a8ce356cd0",
    "id": null,
    "metadata": {},
    "name": "PolicyPageCreateDraftPolicyVersionMutation",
    "operationKind": "mutation",
    "text": "mutation PolicyPageCreateDraftPolicyVersionMutation(\n  $input: CreateDraftPolicyVersionInput!\n) {\n  createDraftPolicyVersion(input: $input) {\n    policyVersionEdge {\n      node {\n        id\n        version\n        status\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7fcf72158c1dc58ae8f9d7efaa6d93a9";

export default node;
