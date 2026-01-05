/**
 * @generated SignedSource<<b8fb35bae12056fe9451875e34aab319>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportTransferImpactAssessmentsPDFInput = {
  filter?: TransferImpactAssessmentFilter | null | undefined;
  organizationId: string;
};
export type TransferImpactAssessmentFilter = {
  snapshotId?: string | null | undefined;
};
export type ProcessingActivitiesPageExportTIAPDFMutation$variables = {
  input: ExportTransferImpactAssessmentsPDFInput;
};
export type ProcessingActivitiesPageExportTIAPDFMutation$data = {
  readonly exportTransferImpactAssessmentsPDF: {
    readonly data: string;
  };
};
export type ProcessingActivitiesPageExportTIAPDFMutation = {
  response: ProcessingActivitiesPageExportTIAPDFMutation$data;
  variables: ProcessingActivitiesPageExportTIAPDFMutation$variables;
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
    "concreteType": "ExportTransferImpactAssessmentsPDFPayload",
    "kind": "LinkedField",
    "name": "exportTransferImpactAssessmentsPDF",
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
    "name": "ProcessingActivitiesPageExportTIAPDFMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivitiesPageExportTIAPDFMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "b41ab5f56ee5dc4436069c4e2a97f8d8",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivitiesPageExportTIAPDFMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivitiesPageExportTIAPDFMutation(\n  $input: ExportTransferImpactAssessmentsPDFInput!\n) {\n  exportTransferImpactAssessmentsPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "3fe289d5564d0617074825d8c58f88f8";

export default node;
