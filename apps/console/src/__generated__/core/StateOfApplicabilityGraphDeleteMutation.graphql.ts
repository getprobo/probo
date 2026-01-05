/**
 * @generated SignedSource<<deb6f38f36dc011d0fc4fc060303a57a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteStateOfApplicabilityInput = {
  stateOfApplicabilityId: string;
};
export type StateOfApplicabilityGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteStateOfApplicabilityInput;
};
export type StateOfApplicabilityGraphDeleteMutation$data = {
  readonly deleteStateOfApplicability: {
    readonly deletedStateOfApplicabilityId: string;
  };
};
export type StateOfApplicabilityGraphDeleteMutation = {
  response: StateOfApplicabilityGraphDeleteMutation$data;
  variables: StateOfApplicabilityGraphDeleteMutation$variables;
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
  "name": "deletedStateOfApplicabilityId",
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
    "name": "StateOfApplicabilityGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteStateOfApplicabilityPayload",
        "kind": "LinkedField",
        "name": "deleteStateOfApplicability",
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
    "name": "StateOfApplicabilityGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteStateOfApplicabilityPayload",
        "kind": "LinkedField",
        "name": "deleteStateOfApplicability",
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
            "name": "deletedStateOfApplicabilityId",
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
    "cacheID": "197a2ed2b48d43372e16deae6e9b98c2",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityGraphDeleteMutation(\n  $input: DeleteStateOfApplicabilityInput!\n) {\n  deleteStateOfApplicability(input: $input) {\n    deletedStateOfApplicabilityId\n  }\n}\n"
  }
};
})();

(node as any).hash = "92ec928d4ad0fe7ff4ae78ad9f1f329a";

export default node;
