/**
 * @generated SignedSource<<ac19789e6716697e6723a0a7b108cb81>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportReportPDFInput = {
  reportId: string;
};
export type AuditRowDownloadMutation$variables = {
  input: ExportReportPDFInput;
};
export type AuditRowDownloadMutation$data = {
  readonly exportReportPDF: {
    readonly data: string;
  };
};
export type AuditRowDownloadMutation = {
  response: AuditRowDownloadMutation$data;
  variables: AuditRowDownloadMutation$variables;
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
    "concreteType": "ExportReportPDFPayload",
    "kind": "LinkedField",
    "name": "exportReportPDF",
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
    "name": "AuditRowDownloadMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AuditRowDownloadMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "79de6a9f2755587818dac20ec392b8e7",
    "id": null,
    "metadata": {},
    "name": "AuditRowDownloadMutation",
    "operationKind": "mutation",
    "text": "mutation AuditRowDownloadMutation(\n  $input: ExportReportPDFInput!\n) {\n  exportReportPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "e3d7ceb7a0da979dad83b1b8289e192e";

export default node;
