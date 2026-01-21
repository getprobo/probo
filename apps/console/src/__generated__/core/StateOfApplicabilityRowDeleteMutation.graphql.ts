/**
 * @generated SignedSource<<74458c1ec95a320a91f42f0c5cdb0925>>
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
export type StateOfApplicabilityRowDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteStateOfApplicabilityInput;
};
export type StateOfApplicabilityRowDeleteMutation$data = {
  readonly deleteStateOfApplicability: {
    readonly deletedStateOfApplicabilityId: string;
  };
};
export type StateOfApplicabilityRowDeleteMutation = {
  response: StateOfApplicabilityRowDeleteMutation$data;
  variables: StateOfApplicabilityRowDeleteMutation$variables;
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
    "name": "StateOfApplicabilityRowDeleteMutation",
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
    "name": "StateOfApplicabilityRowDeleteMutation",
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
    "cacheID": "be77dc47292c3864174f008adb45f910",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityRowDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityRowDeleteMutation(\n  $input: DeleteStateOfApplicabilityInput!\n) {\n  deleteStateOfApplicability(input: $input) {\n    deletedStateOfApplicabilityId\n  }\n}\n"
  }
};
})();

(node as any).hash = "c580dcff6800cb372932ec5ed9201417";

export default node;
