/**
 * @generated SignedSource<<02f25af2f023ddc0a964b44c105ea248>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportDocumentVersionPDFInput = {
  documentVersionId: string;
  watermarkEmail?: string | null | undefined;
  withSignatures: boolean;
  withWatermark: boolean;
};
export type DocumentActionsDropdownn_exportVersionMutation$variables = {
  input: ExportDocumentVersionPDFInput;
};
export type DocumentActionsDropdownn_exportVersionMutation$data = {
  readonly exportDocumentVersionPDF: {
    readonly data: string;
  };
};
export type DocumentActionsDropdownn_exportVersionMutation = {
  response: DocumentActionsDropdownn_exportVersionMutation$data;
  variables: DocumentActionsDropdownn_exportVersionMutation$variables;
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
    "concreteType": "ExportDocumentVersionPDFPayload",
    "kind": "LinkedField",
    "name": "exportDocumentVersionPDF",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "data",
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
    "name": "DocumentActionsDropdownn_exportVersionMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentActionsDropdownn_exportVersionMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "b9228ff1ce940910b277cecf021ebecb",
    "id": null,
    "metadata": {},
    "name": "DocumentActionsDropdownn_exportVersionMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentActionsDropdownn_exportVersionMutation(\n  $input: ExportDocumentVersionPDFInput!\n) {\n  exportDocumentVersionPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "1bab1476f7c4231736feefd227063c27";

export default node;
