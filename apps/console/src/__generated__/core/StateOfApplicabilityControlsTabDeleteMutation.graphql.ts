/**
 * @generated SignedSource<<efea63fc9ceccad58af05da557d0ac53>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteApplicabilityStatementInput = {
  applicabilityStatementId: string;
};
export type StateOfApplicabilityControlsTabDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteApplicabilityStatementInput;
};
export type StateOfApplicabilityControlsTabDeleteMutation$data = {
  readonly deleteApplicabilityStatement: {
    readonly deletedApplicabilityStatementId: string;
  };
};
export type StateOfApplicabilityControlsTabDeleteMutation = {
  response: StateOfApplicabilityControlsTabDeleteMutation$data;
  variables: StateOfApplicabilityControlsTabDeleteMutation$variables;
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
  "name": "deletedApplicabilityStatementId",
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
    "name": "StateOfApplicabilityControlsTabDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "deleteApplicabilityStatement",
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
    "name": "StateOfApplicabilityControlsTabDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteApplicabilityStatementPayload",
        "kind": "LinkedField",
        "name": "deleteApplicabilityStatement",
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
            "name": "deletedApplicabilityStatementId",
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
    "cacheID": "427051073f5e747e828d282a773e4db6",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityControlsTabDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityControlsTabDeleteMutation(\n  $input: DeleteApplicabilityStatementInput!\n) {\n  deleteApplicabilityStatement(input: $input) {\n    deletedApplicabilityStatementId\n  }\n}\n"
  }
};
})();

(node as any).hash = "860e6ba50b4d8cfd9dd53416d5e3b710";

export default node;
