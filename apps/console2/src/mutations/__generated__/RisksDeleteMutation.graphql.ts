/**
 * @generated SignedSource<<5102c0f9ea660855dc5db7c6b16fd64b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteRiskInput = {
  riskId: string;
};
export type RisksDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteRiskInput;
};
export type RisksDeleteMutation$data = {
  readonly deleteRisk: {
    readonly deletedRiskId: string;
  };
};
export type RisksDeleteMutation = {
  response: RisksDeleteMutation$data;
  variables: RisksDeleteMutation$variables;
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
  "name": "deletedRiskId",
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
    "name": "RisksDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRiskPayload",
        "kind": "LinkedField",
        "name": "deleteRisk",
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
    "name": "RisksDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRiskPayload",
        "kind": "LinkedField",
        "name": "deleteRisk",
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
            "name": "deletedRiskId",
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
    "cacheID": "f4273ec8d2cd76543169a6d1b7cd72b3",
    "id": null,
    "metadata": {},
    "name": "RisksDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation RisksDeleteMutation(\n  $input: DeleteRiskInput!\n) {\n  deleteRisk(input: $input) {\n    deletedRiskId\n  }\n}\n"
  }
};
})();

(node as any).hash = "ad609ad0f9c749fe47f2bd0d382777bb";

export default node;
