/**
 * @generated SignedSource<<9dc238538172b956fb81954024522ee1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportStateOfApplicabilityPDFInput = {
  stateOfApplicabilityId: string;
};
export type StateOfApplicabilityDetailPageExportMutation$variables = {
  input: ExportStateOfApplicabilityPDFInput;
};
export type StateOfApplicabilityDetailPageExportMutation$data = {
  readonly exportStateOfApplicabilityPDF: {
    readonly data: string;
  };
};
export type StateOfApplicabilityDetailPageExportMutation = {
  response: StateOfApplicabilityDetailPageExportMutation$data;
  variables: StateOfApplicabilityDetailPageExportMutation$variables;
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
    "concreteType": "ExportStateOfApplicabilityPDFPayload",
    "kind": "LinkedField",
    "name": "exportStateOfApplicabilityPDF",
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
    "name": "StateOfApplicabilityDetailPageExportMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "StateOfApplicabilityDetailPageExportMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "dbc24d23ab71abb76f95b114f86f200d",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityDetailPageExportMutation",
    "operationKind": "mutation",
    "text": "mutation StateOfApplicabilityDetailPageExportMutation(\n  $input: ExportStateOfApplicabilityPDFInput!\n) {\n  exportStateOfApplicabilityPDF(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "36be308a451ada118c2bcc9c3b0c0a2b";

export default node;
