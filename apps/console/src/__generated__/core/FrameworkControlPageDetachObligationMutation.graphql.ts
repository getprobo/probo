/**
 * @generated SignedSource<<b1b8d1f4c0dc579cacc0434be9b36259>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlObligationMappingInput = {
  controlId: string;
  obligationId: string;
};
export type FrameworkControlPageDetachObligationMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlObligationMappingInput;
};
export type FrameworkControlPageDetachObligationMutation$data = {
  readonly deleteControlObligationMapping: {
    readonly deletedObligationId: string;
  };
};
export type FrameworkControlPageDetachObligationMutation = {
  response: FrameworkControlPageDetachObligationMutation$data;
  variables: FrameworkControlPageDetachObligationMutation$variables;
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
  "name": "deletedObligationId",
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
    "name": "FrameworkControlPageDetachObligationMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlObligationMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlObligationMapping",
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
    "name": "FrameworkControlPageDetachObligationMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlObligationMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlObligationMapping",
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
            "name": "deletedObligationId",
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
    "cacheID": "90f4a67d3a22217cc62841266c5a0450",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDetachObligationMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDetachObligationMutation(\n  $input: DeleteControlObligationMappingInput!\n) {\n  deleteControlObligationMapping(input: $input) {\n    deletedObligationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "92da0169c3d421a658aa560b4e5ba1c8";

export default node;
