/**
 * @generated SignedSource<<82ece1ab5e9585bfcce605e38fe9563a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportProcessingActivitiesPDFInput = {
  filter?: ProcessingActivityFilter | null | undefined;
  organizationId: string;
};
export type ProcessingActivityFilter = {
  snapshotId?: string | null | undefined;
};
export type ProcessingActivitiesPageExportPDFMutation$variables = {
  input: ExportProcessingActivitiesPDFInput;
};
export type ProcessingActivitiesPageExportPDFMutation$data = {
  readonly exportProcessingActivitiesPDF: {
    readonly data: string;
  };
};
export type ProcessingActivitiesPageExportPDFMutation = {
  response: ProcessingActivitiesPageExportPDFMutation$data;
  variables: ProcessingActivitiesPageExportPDFMutation$variables;
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
    "concreteType": "ExportProcessingActivitiesPDFPayload",
    "kind": "LinkedField",
    "name": "exportProcessingActivitiesPDF",
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
    "name": "ProcessingActivitiesPageExportPDFMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivitiesPageExportPDFMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "85140bef5e55c8ae4fdb139965375f13",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivitiesPageExportPDFMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivitiesPageExportPDFMutation(\n  $input: ExportProcessingActivitiesPDFInput!\n) {\n  exportProcessingActivitiesPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "cc777bb4440f6aed36c2fb92c1bc7e0b";

export default node;
