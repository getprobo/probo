/**
 * @generated SignedSource<<86e0627df7e535b4ca603724aa4230e8>>
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
export type useDeleteStateOfApplicabilityMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteStateOfApplicabilityInput;
};
export type useDeleteStateOfApplicabilityMutation$data = {
  readonly deleteStateOfApplicability: {
    readonly deletedStateOfApplicabilityId: string;
  };
};
export type useDeleteStateOfApplicabilityMutation = {
  response: useDeleteStateOfApplicabilityMutation$data;
  variables: useDeleteStateOfApplicabilityMutation$variables;
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
    "name": "useDeleteStateOfApplicabilityMutation",
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
    "name": "useDeleteStateOfApplicabilityMutation",
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
    "cacheID": "d828489f1c84fa96748d9f723bec1c2f",
    "id": null,
    "metadata": {},
    "name": "useDeleteStateOfApplicabilityMutation",
    "operationKind": "mutation",
    "text": "mutation useDeleteStateOfApplicabilityMutation(\n  $input: DeleteStateOfApplicabilityInput!\n) {\n  deleteStateOfApplicability(input: $input) {\n    deletedStateOfApplicabilityId\n  }\n}\n"
  }
};
})();

(node as any).hash = "848e21cc5d5ed2e7d65b7eb0e3413c68";

export default node;
