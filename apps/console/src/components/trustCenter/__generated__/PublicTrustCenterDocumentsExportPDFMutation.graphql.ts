/**
 * @generated SignedSource<<24fd0dcf15c98ad3d20762e4ccc83a1b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportDocumentVersionPDFInput = {
  documentVersionId: string;
};
export type PublicTrustCenterDocumentsExportPDFMutation$variables = {
  input: ExportDocumentVersionPDFInput;
};
export type PublicTrustCenterDocumentsExportPDFMutation$data = {
  readonly exportDocumentVersionPDF: {
    readonly data: string;
  };
};
export type PublicTrustCenterDocumentsExportPDFMutation = {
  response: PublicTrustCenterDocumentsExportPDFMutation$data;
  variables: PublicTrustCenterDocumentsExportPDFMutation$variables;
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
    "name": "PublicTrustCenterDocumentsExportPDFMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PublicTrustCenterDocumentsExportPDFMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "b5d33d4c301cfa811bb74d52f61f52e8",
    "id": null,
    "metadata": {},
    "name": "PublicTrustCenterDocumentsExportPDFMutation",
    "operationKind": "mutation",
    "text": "mutation PublicTrustCenterDocumentsExportPDFMutation(\n  $input: ExportDocumentVersionPDFInput!\n) {\n  exportDocumentVersionPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "b6a173895aea2450ad4ac1a4ec6aeb4e";

export default node;
