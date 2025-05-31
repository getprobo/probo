/**
 * @generated SignedSource<<c68450dee24a94b20d5c4985a83511e0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PublishPolicyVersionInput = {
  policyId: string;
};
export type PolicyPagePublishMutation$variables = {
  input: PublishPolicyVersionInput;
};
export type PolicyPagePublishMutation$data = {
  readonly publishPolicyVersion: {
    readonly policy: {
      readonly id: string;
    };
  };
};
export type PolicyPagePublishMutation = {
  response: PolicyPagePublishMutation$data;
  variables: PolicyPagePublishMutation$variables;
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
    "concreteType": "PublishPolicyVersionPayload",
    "kind": "LinkedField",
    "name": "publishPolicyVersion",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Policy",
        "kind": "LinkedField",
        "name": "policy",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
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
    "name": "PolicyPagePublishMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PolicyPagePublishMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "4279dd7a9a12d8d5fb6e07c1875de00a",
    "id": null,
    "metadata": {},
    "name": "PolicyPagePublishMutation",
    "operationKind": "mutation",
    "text": "mutation PolicyPagePublishMutation(\n  $input: PublishPolicyVersionInput!\n) {\n  publishPolicyVersion(input: $input) {\n    policy {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8ee6cdf68342db34ec04aab82313eff0";

export default node;
