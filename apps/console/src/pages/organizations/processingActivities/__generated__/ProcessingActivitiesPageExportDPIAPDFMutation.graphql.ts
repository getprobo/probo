/**
 * @generated SignedSource<<2f43f6874d8b6d515cf244934b7b19ea>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportDataProtectionImpactAssessmentsPDFInput = {
  filter?: DataProtectionImpactAssessmentFilter | null | undefined;
  organizationId: string;
};
export type DataProtectionImpactAssessmentFilter = {
  snapshotId?: string | null | undefined;
};
export type ProcessingActivitiesPageExportDPIAPDFMutation$variables = {
  input: ExportDataProtectionImpactAssessmentsPDFInput;
};
export type ProcessingActivitiesPageExportDPIAPDFMutation$data = {
  readonly exportDataProtectionImpactAssessmentsPDF: {
    readonly data: string;
  };
};
export type ProcessingActivitiesPageExportDPIAPDFMutation = {
  response: ProcessingActivitiesPageExportDPIAPDFMutation$data;
  variables: ProcessingActivitiesPageExportDPIAPDFMutation$variables;
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
    "concreteType": "ExportDataProtectionImpactAssessmentsPDFPayload",
    "kind": "LinkedField",
    "name": "exportDataProtectionImpactAssessmentsPDF",
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
    "name": "ProcessingActivitiesPageExportDPIAPDFMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivitiesPageExportDPIAPDFMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "851a1c0def7719185cf3a67954696231",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivitiesPageExportDPIAPDFMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivitiesPageExportDPIAPDFMutation(\n  $input: ExportDataProtectionImpactAssessmentsPDFInput!\n) {\n  exportDataProtectionImpactAssessmentsPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "3c2a96fb29c61791e7df39c7b8174c90";

export default node;
