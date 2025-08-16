/**
 * @generated SignedSource<<eb0bdb8e311cea402d1aa11b79c2c1a5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteNonconformityRegistryInput = {
  nonconformityRegistryId: string;
};
export type NonconformityRegistryGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteNonconformityRegistryInput;
};
export type NonconformityRegistryGraphDeleteMutation$data = {
  readonly deleteNonconformityRegistry: {
    readonly deletedNonconformityRegistryId: string;
  };
};
export type NonconformityRegistryGraphDeleteMutation = {
  response: NonconformityRegistryGraphDeleteMutation$data;
  variables: NonconformityRegistryGraphDeleteMutation$variables;
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
  "name": "deletedNonconformityRegistryId",
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
    "name": "NonconformityRegistryGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteNonconformityRegistryPayload",
        "kind": "LinkedField",
        "name": "deleteNonconformityRegistry",
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
    "name": "NonconformityRegistryGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteNonconformityRegistryPayload",
        "kind": "LinkedField",
        "name": "deleteNonconformityRegistry",
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
            "name": "deletedNonconformityRegistryId",
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
    "cacheID": "733821706945a20396447b1134615a0e",
    "id": null,
    "metadata": {},
    "name": "NonconformityRegistryGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation NonconformityRegistryGraphDeleteMutation(\n  $input: DeleteNonconformityRegistryInput!\n) {\n  deleteNonconformityRegistry(input: $input) {\n    deletedNonconformityRegistryId\n  }\n}\n"
  }
};
})();

(node as any).hash = "99bc5c0c79ffa8de365ab541c960fc84";

export default node;
