/**
 * @generated SignedSource<<baab5b247999631abac20d0f5ab13925>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeletePolicyInput = {
  policyId: string;
};
export type PolicyGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeletePolicyInput;
};
export type PolicyGraphDeleteMutation$data = {
  readonly deletePolicy: {
    readonly deletedPolicyId: string;
  };
};
export type PolicyGraphDeleteMutation = {
  response: PolicyGraphDeleteMutation$data;
  variables: PolicyGraphDeleteMutation$variables;
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
  "name": "deletedPolicyId",
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
    "name": "PolicyGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeletePolicyPayload",
        "kind": "LinkedField",
        "name": "deletePolicy",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "PolicyGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeletePolicyPayload",
        "kind": "LinkedField",
        "name": "deletePolicy",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedPolicyId",
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
    "cacheID": "d635ed7113bd8ad32056f5b4fbb95b22",
    "id": null,
    "metadata": {},
    "name": "PolicyGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation PolicyGraphDeleteMutation(\n  $input: DeletePolicyInput!\n) {\n  deletePolicy(input: $input) {\n    deletedPolicyId\n  }\n}\n"
  }
};
})();

(node as any).hash = "f5178185f41864d20506abc44b37fa67";

export default node;
