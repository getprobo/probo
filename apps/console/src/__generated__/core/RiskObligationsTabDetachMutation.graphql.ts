/**
 * @generated SignedSource<<58dd326af2caac72cf50ed8f289a3cac>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteRiskObligationMappingInput = {
  obligationId: string;
  riskId: string;
};
export type RiskObligationsTabDetachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteRiskObligationMappingInput;
};
export type RiskObligationsTabDetachMutation$data = {
  readonly deleteRiskObligationMapping: {
    readonly deletedObligationId: string;
  };
};
export type RiskObligationsTabDetachMutation = {
  response: RiskObligationsTabDetachMutation$data;
  variables: RiskObligationsTabDetachMutation$variables;
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
    "name": "RiskObligationsTabDetachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRiskObligationMappingPayload",
        "kind": "LinkedField",
        "name": "deleteRiskObligationMapping",
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
    "name": "RiskObligationsTabDetachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRiskObligationMappingPayload",
        "kind": "LinkedField",
        "name": "deleteRiskObligationMapping",
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
    "cacheID": "97b75f003c5e9518831c9699a2fcaff0",
    "id": null,
    "metadata": {},
    "name": "RiskObligationsTabDetachMutation",
    "operationKind": "mutation",
    "text": "mutation RiskObligationsTabDetachMutation(\n  $input: DeleteRiskObligationMappingInput!\n) {\n  deleteRiskObligationMapping(input: $input) {\n    deletedObligationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "5923c7e9bcb08224ba1a60b0ad09642a";

export default node;
