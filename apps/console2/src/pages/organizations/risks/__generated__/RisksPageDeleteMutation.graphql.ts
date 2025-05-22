/**
 * @generated SignedSource<<b410e2d0b99401f7ad86c1d4f7a0fcfe>>
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
export type RisksPageDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteRiskInput;
};
export type RisksPageDeleteMutation$data = {
  readonly deleteRisk: {
    readonly deletedRiskId: string;
  };
};
export type RisksPageDeleteMutation = {
  response: RisksPageDeleteMutation$data;
  variables: RisksPageDeleteMutation$variables;
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
    "name": "RisksPageDeleteMutation",
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
    "name": "RisksPageDeleteMutation",
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
    "cacheID": "a08977dc1d81b5ab63e431014b2d2b41",
    "id": null,
    "metadata": {},
    "name": "RisksPageDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation RisksPageDeleteMutation(\n  $input: DeleteRiskInput!\n) {\n  deleteRisk(input: $input) {\n    deletedRiskId\n  }\n}\n"
  }
};
})();

(node as any).hash = "fb17a17779665063283d3d73fcb39785";

export default node;
