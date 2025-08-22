/**
 * @generated SignedSource<<e737361b21a8767acd6a604702875029>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportDocumentPDFInput = {
  documentId: string;
};
export type PublicTrustCenterDocumentsExportPDFMutation$variables = {
  input: ExportDocumentPDFInput;
};
export type PublicTrustCenterDocumentsExportPDFMutation$data = {
  readonly exportDocumentPDF: {
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
    "concreteType": "ExportDocumentPDFPayload",
    "kind": "LinkedField",
    "name": "exportDocumentPDF",
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
    "cacheID": "bb7c09ea21c22c0a728b18dde2449f72",
    "id": null,
    "metadata": {},
    "name": "PublicTrustCenterDocumentsExportPDFMutation",
    "operationKind": "mutation",
    "text": "mutation PublicTrustCenterDocumentsExportPDFMutation(\n  $input: ExportDocumentPDFInput!\n) {\n  exportDocumentPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "00a3f3d260fe38b35fe33a0033cf2400";

export default node;
