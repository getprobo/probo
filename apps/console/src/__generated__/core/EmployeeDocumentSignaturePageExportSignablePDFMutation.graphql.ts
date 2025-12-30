/**
 * @generated SignedSource<<bd62735d2891123ff7f79b1b473d5143>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportSignableDocumentVersionPDFInput = {
  documentVersionId: string;
};
export type EmployeeDocumentSignaturePageExportSignablePDFMutation$variables = {
  input: ExportSignableDocumentVersionPDFInput;
};
export type EmployeeDocumentSignaturePageExportSignablePDFMutation$data = {
  readonly exportSignableVersionDocumentPDF: {
    readonly data: string;
  };
};
export type EmployeeDocumentSignaturePageExportSignablePDFMutation = {
  response: EmployeeDocumentSignaturePageExportSignablePDFMutation$data;
  variables: EmployeeDocumentSignaturePageExportSignablePDFMutation$variables;
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
    "concreteType": "ExportSignableDocumentVersionPDFPayload",
    "kind": "LinkedField",
    "name": "exportSignableVersionDocumentPDF",
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
    "name": "EmployeeDocumentSignaturePageExportSignablePDFMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EmployeeDocumentSignaturePageExportSignablePDFMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "1bbeaaac843ecd06e9f7ee662aa11fc2",
    "id": null,
    "metadata": {},
    "name": "EmployeeDocumentSignaturePageExportSignablePDFMutation",
    "operationKind": "mutation",
    "text": "mutation EmployeeDocumentSignaturePageExportSignablePDFMutation(\n  $input: ExportSignableDocumentVersionPDFInput!\n) {\n  exportSignableVersionDocumentPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "39a1e34d4b4f8c98d262dd3a737ebb7c";

export default node;
