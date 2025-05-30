/**
 * @generated SignedSource<<77d14d5c3a949757604407daf7ba2600>>
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
export type ShowRiskViewDeleteRiskDocumentMappingMutation$variables = {
  input: DeleteRiskDocumentMappingInput;
};
export type ShowRiskViewDeleteRiskDocumentMappingMutation$data = {
  readonly deleteRiskDocumentMapping: {
    readonly deletedDocumentId: string;
  };
};
export type ShowRiskViewDeleteRiskDocumentMappingMutation = {
  response: ShowRiskViewDeleteRiskDocumentMappingMutation$data;
  variables: ShowRiskViewDeleteRiskDocumentMappingMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "DeleteRiskDocumentMappingPayload",
    "kind": "LinkedField",
    "name": "deleteRiskDocumentMapping",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedDocumentId",
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ShowRiskViewDeleteRiskDocumentMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ShowRiskViewDeleteRiskDocumentMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "1cd0845f63cf338aed2c2e565e239306",
    "id": null,
    "metadata": {},
    "name": "ShowRiskViewDeleteRiskDocumentMappingMutation",
    "operationKind": "mutation",
    "text": "mutation ShowRiskViewDeleteRiskDocumentMappingMutation(\n  $input: DeleteRiskDocumentMappingInput!\n) {\n  deleteRiskDocumentMapping(input: $input) {\n    deletedDocumentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "b7b61daef6f0544c6b78f69094ad2d82";

export default node;
