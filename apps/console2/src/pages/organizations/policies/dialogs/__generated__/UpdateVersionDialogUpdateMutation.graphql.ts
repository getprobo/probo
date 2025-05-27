/**
 * @generated SignedSource<<6cbec0c0b675969ecdebebf74c7465ab>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdatePolicyVersionInput = {
  content: string;
  policyVersionId: string;
};
export type UpdateVersionDialogUpdateMutation$variables = {
  input: UpdatePolicyVersionInput;
};
export type UpdateVersionDialogUpdateMutation$data = {
  readonly updatePolicyVersion: {
    readonly policyVersion: {
      readonly content: string;
      readonly id: string;
    };
  };
};
export type UpdateVersionDialogUpdateMutation = {
  response: UpdateVersionDialogUpdateMutation$data;
  variables: UpdateVersionDialogUpdateMutation$variables;
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
    "concreteType": "UpdatePolicyVersionPayload",
    "kind": "LinkedField",
    "name": "updatePolicyVersion",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "PolicyVersion",
        "kind": "LinkedField",
        "name": "policyVersion",
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
            "name": "content",
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
    "name": "UpdateVersionDialogUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "UpdateVersionDialogUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "374557bf1319583b7d590f5bea438e63",
    "id": null,
    "metadata": {},
    "name": "UpdateVersionDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation UpdateVersionDialogUpdateMutation(\n  $input: UpdatePolicyVersionInput!\n) {\n  updatePolicyVersion(input: $input) {\n    policyVersion {\n      id\n      content\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "d863d893bfa969b02768096e024275a4";

export default node;
