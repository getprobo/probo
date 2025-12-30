/**
 * @generated SignedSource<<3a082179d9ea9d47824b7ee925418b95>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportDocumentVersionPDFInput = {
  documentVersionId: string;
  watermarkEmail?: any | null | undefined;
  withSignatures: boolean;
  withWatermark: boolean;
};
export type DocumentDetailPageExportPDFMutation$variables = {
  input: ExportDocumentVersionPDFInput;
};
export type DocumentDetailPageExportPDFMutation$data = {
  readonly exportDocumentVersionPDF: {
    readonly data: string;
  };
};
export type DocumentDetailPageExportPDFMutation = {
  response: DocumentDetailPageExportPDFMutation$data;
  variables: DocumentDetailPageExportPDFMutation$variables;
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
    "name": "DocumentDetailPageExportPDFMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentDetailPageExportPDFMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "4c17d48a790ff5f1617c7db604766b83",
    "id": null,
    "metadata": {},
    "name": "DocumentDetailPageExportPDFMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentDetailPageExportPDFMutation(\n  $input: ExportDocumentVersionPDFInput!\n) {\n  exportDocumentVersionPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "e8854650389c3ebd1020ec278d8be831";

export default node;
