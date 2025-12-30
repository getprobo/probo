/**
 * @generated SignedSource<<ae02449f816da7f2b79f16edbe2f7066>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteRiskDocumentMappingInput = {
  documentId: string;
  riskId: string;
};
export type RiskDocumentsTabDetachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteRiskDocumentMappingInput;
};
export type RiskDocumentsTabDetachMutation$data = {
  readonly deleteRiskDocumentMapping: {
    readonly deletedDocumentId: string;
  };
};
export type RiskDocumentsTabDetachMutation = {
  response: RiskDocumentsTabDetachMutation$data;
  variables: RiskDocumentsTabDetachMutation$variables;
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
  "name": "deletedDocumentId",
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
    "name": "RiskDocumentsTabDetachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRiskDocumentMappingPayload",
        "kind": "LinkedField",
        "name": "deleteRiskDocumentMapping",
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
    "name": "RiskDocumentsTabDetachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteRiskDocumentMappingPayload",
        "kind": "LinkedField",
        "name": "deleteRiskDocumentMapping",
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
            "name": "deletedDocumentId",
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
    "cacheID": "a8dc5064c650c498193904e16790df6e",
    "id": null,
    "metadata": {},
    "name": "RiskDocumentsTabDetachMutation",
    "operationKind": "mutation",
    "text": "mutation RiskDocumentsTabDetachMutation(\n  $input: DeleteRiskDocumentMappingInput!\n) {\n  deleteRiskDocumentMapping(input: $input) {\n    deletedDocumentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "0aa4bf52197d033dfc3aabca9288aad4";

export default node;
