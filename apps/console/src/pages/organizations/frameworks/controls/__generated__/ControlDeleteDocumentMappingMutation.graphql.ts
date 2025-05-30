/**
 * @generated SignedSource<<a39e2bd50fb3d78ac40f96740496fcac>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlDocumentMappingInput = {
  controlId: string;
  documentId: string;
};
export type ControlDeleteDocumentMappingMutation$variables = {
  input: DeleteControlDocumentMappingInput;
};
export type ControlDeleteDocumentMappingMutation$data = {
  readonly deleteControlDocumentMapping: {
    readonly deletedDocumentId: string;
  };
};
export type ControlDeleteDocumentMappingMutation = {
  response: ControlDeleteDocumentMappingMutation$data;
  variables: ControlDeleteDocumentMappingMutation$variables;
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
    "concreteType": "DeleteControlDocumentMappingPayload",
    "kind": "LinkedField",
    "name": "deleteControlDocumentMapping",
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
    "name": "ControlDeleteDocumentMappingMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ControlDeleteDocumentMappingMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "de5f0d5a1db8bf370f893e99e58729df",
    "id": null,
    "metadata": {},
    "name": "ControlDeleteDocumentMappingMutation",
    "operationKind": "mutation",
    "text": "mutation ControlDeleteDocumentMappingMutation(\n  $input: DeleteControlDocumentMappingInput!\n) {\n  deleteControlDocumentMapping(input: $input) {\n    deletedDocumentId\n  }\n}\n"
  }
};
})();

(node as any).hash = "d7e9d7d588053130899fe2f52399a99d";

export default node;
