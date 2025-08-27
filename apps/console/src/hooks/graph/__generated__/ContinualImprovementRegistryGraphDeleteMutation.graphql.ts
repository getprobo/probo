/**
 * @generated SignedSource<<55e8b0de5cb20c0a479bb4862aab4677>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteContinualImprovementRegistryInput = {
  continualImprovementRegistryId: string;
};
export type ContinualImprovementRegistryGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteContinualImprovementRegistryInput;
};
export type ContinualImprovementRegistryGraphDeleteMutation$data = {
  readonly deleteContinualImprovementRegistry: {
    readonly deletedContinualImprovementRegistryId: string;
  };
};
export type ContinualImprovementRegistryGraphDeleteMutation = {
  response: ContinualImprovementRegistryGraphDeleteMutation$data;
  variables: ContinualImprovementRegistryGraphDeleteMutation$variables;
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
  "name": "deletedContinualImprovementRegistryId",
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
    "name": "ContinualImprovementRegistryGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteContinualImprovementRegistryPayload",
        "kind": "LinkedField",
        "name": "deleteContinualImprovementRegistry",
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
    "name": "ContinualImprovementRegistryGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteContinualImprovementRegistryPayload",
        "kind": "LinkedField",
        "name": "deleteContinualImprovementRegistry",
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
            "name": "deletedContinualImprovementRegistryId",
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
    "cacheID": "1f973509e72c6a1a729e98a8b4240b25",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementRegistryGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementRegistryGraphDeleteMutation(\n  $input: DeleteContinualImprovementRegistryInput!\n) {\n  deleteContinualImprovementRegistry(input: $input) {\n    deletedContinualImprovementRegistryId\n  }\n}\n"
  }
};
})();

(node as any).hash = "7c7512a398e2391d5f7f083395cc6395";

export default node;
