/**
 * @generated SignedSource<<cae4044b1357cd4510e12f83f19001ba>>
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
export type DocumentRowDownloadMutation$variables = {
  input: ExportDocumentPDFInput;
};
export type DocumentRowDownloadMutation$data = {
  readonly exportDocumentPDF: {
    readonly data: string;
  };
};
export type DocumentRowDownloadMutation = {
  response: DocumentRowDownloadMutation$data;
  variables: DocumentRowDownloadMutation$variables;
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
    "name": "DocumentRowDownloadMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentRowDownloadMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "126681ef86a7a6aa5c831bdf53b7cf3e",
    "id": null,
    "metadata": {},
    "name": "DocumentRowDownloadMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentRowDownloadMutation(\n  $input: ExportDocumentPDFInput!\n) {\n  exportDocumentPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "8e48a364a7f9cf3d3a51a9db60ff792f";

export default node;
