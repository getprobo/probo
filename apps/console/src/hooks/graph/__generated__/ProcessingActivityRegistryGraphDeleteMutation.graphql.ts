/**
 * @generated SignedSource<<11633889a4d61df894b88dd483687cf9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteProcessingActivityRegistryInput = {
  processingActivityRegistryId: string;
};
export type ProcessingActivityRegistryGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteProcessingActivityRegistryInput;
};
export type ProcessingActivityRegistryGraphDeleteMutation$data = {
  readonly deleteProcessingActivityRegistry: {
    readonly deletedProcessingActivityRegistryId: string;
  };
};
export type ProcessingActivityRegistryGraphDeleteMutation = {
  response: ProcessingActivityRegistryGraphDeleteMutation$data;
  variables: ProcessingActivityRegistryGraphDeleteMutation$variables;
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
  "name": "deletedProcessingActivityRegistryId",
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
    "name": "ProcessingActivityRegistryGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteProcessingActivityRegistryPayload",
        "kind": "LinkedField",
        "name": "deleteProcessingActivityRegistry",
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
    "name": "ProcessingActivityRegistryGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteProcessingActivityRegistryPayload",
        "kind": "LinkedField",
        "name": "deleteProcessingActivityRegistry",
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
            "name": "deletedProcessingActivityRegistryId",
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
    "cacheID": "7550d138c9636bbe930a5cdad260156b",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityRegistryGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityRegistryGraphDeleteMutation(\n  $input: DeleteProcessingActivityRegistryInput!\n) {\n  deleteProcessingActivityRegistry(input: $input) {\n    deletedProcessingActivityRegistryId\n  }\n}\n"
  }
};
})();

(node as any).hash = "8492e778e152e5910ee4c09c34ea4e68";

export default node;
