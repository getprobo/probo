/**
 * @generated SignedSource<<26ab8c31c2ed501581810e36579da08b>>
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
export type ReportRowDownloadMutation$variables = {
  input: ExportReportPDFInput;
};
export type ReportRowDownloadMutation$data = {
  readonly exportReportPDF: {
    readonly data: string;
  };
};
export type ReportRowDownloadMutation = {
  response: ReportRowDownloadMutation$data;
  variables: ReportRowDownloadMutation$variables;
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
    "name": "ReportRowDownloadMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ReportRowDownloadMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "07784b7e72dd719d1b959c60594e0af7",
    "id": null,
    "metadata": {},
    "name": "ReportRowDownloadMutation",
    "operationKind": "mutation",
    "text": "mutation ReportRowDownloadMutation(\n  $input: ExportReportPDFInput!\n) {\n  exportReportPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "cc3ee76f695432d964d34f05d4a078f5";

export default node;
