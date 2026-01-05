/**
 * @generated SignedSource<<b5ac3d2a6151aa858185089d6d979a34>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteStateOfApplicabilityControlMappingInput = {
  controlId: string;
  stateOfApplicabilityId: string;
};
export type FrameworkControlPageDetachStateOfApplicabilityMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteStateOfApplicabilityControlMappingInput;
};
export type FrameworkControlPageDetachStateOfApplicabilityMutation$data = {
  readonly deleteStateOfApplicabilityControlMapping: {
    readonly deletedControlId: string;
    readonly deletedStateOfApplicabilityControlId: string;
    readonly deletedStateOfApplicabilityId: string;
  };
};
export type FrameworkControlPageDetachStateOfApplicabilityMutation = {
  response: FrameworkControlPageDetachStateOfApplicabilityMutation$data;
  variables: FrameworkControlPageDetachStateOfApplicabilityMutation$variables;
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
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedControlId",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedStateOfApplicabilityControlId",
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
    "name": "FrameworkControlPageDetachStateOfApplicabilityMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteStateOfApplicabilityControlMappingPayload",
        "kind": "LinkedField",
        "name": "deleteStateOfApplicabilityControlMapping",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          (v4/*: any*/),
          (v5/*: any*/)
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
    "name": "FrameworkControlPageDetachStateOfApplicabilityMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteStateOfApplicabilityControlMappingPayload",
        "kind": "LinkedField",
        "name": "deleteStateOfApplicabilityControlMapping",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          (v4/*: any*/),
          (v5/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedStateOfApplicabilityControlId",
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
    "cacheID": "f0ec8162516546b2cdd221c0c9832908",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDetachStateOfApplicabilityMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDetachStateOfApplicabilityMutation(\n  $input: DeleteStateOfApplicabilityControlMappingInput!\n) {\n  deleteStateOfApplicabilityControlMapping(input: $input) {\n    deletedStateOfApplicabilityId\n    deletedControlId\n    deletedStateOfApplicabilityControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "f1a4efbd039ed9f457905903df50a160";

export default node;
