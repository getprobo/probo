/**
 * @generated SignedSource<<7e00cc82c493f1f051c9673fb95c6217>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteContinualImprovementInput = {
  continualImprovementId: string;
};
export type ContinualImprovementGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteContinualImprovementInput;
};
export type ContinualImprovementGraphDeleteMutation$data = {
  readonly deleteContinualImprovement: {
    readonly deletedContinualImprovementId: string;
  };
};
export type ContinualImprovementGraphDeleteMutation = {
  response: ContinualImprovementGraphDeleteMutation$data;
  variables: ContinualImprovementGraphDeleteMutation$variables;
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
  "name": "deletedContinualImprovementId",
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
    "name": "ContinualImprovementGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteContinualImprovementPayload",
        "kind": "LinkedField",
        "name": "deleteContinualImprovement",
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
    "name": "ContinualImprovementGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteContinualImprovementPayload",
        "kind": "LinkedField",
        "name": "deleteContinualImprovement",
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
            "name": "deletedContinualImprovementId",
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
    "cacheID": "aaa3d35494c0b1bcaca143cdbc6b7143",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementGraphDeleteMutation(\n  $input: DeleteContinualImprovementInput!\n) {\n  deleteContinualImprovement(input: $input) {\n    deletedContinualImprovementId\n  }\n}\n"
  }
};
})();

(node as any).hash = "5474ad4a9bdc5ee4bd70f7289944079e";

export default node;
