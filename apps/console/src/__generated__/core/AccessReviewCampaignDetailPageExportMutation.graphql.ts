/**
 * @generated SignedSource<<3f882bd0e0d2b96b497c565349d0941e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportCampaignEvidenceInput = {
  accessReviewCampaignId: string;
};
export type AccessReviewCampaignDetailPageExportMutation$variables = {
  input: ExportCampaignEvidenceInput;
};
export type AccessReviewCampaignDetailPageExportMutation$data = {
  readonly exportCampaignEvidence: {
    readonly checksumSha256: string;
    readonly payload: string;
  };
};
export type AccessReviewCampaignDetailPageExportMutation = {
  response: AccessReviewCampaignDetailPageExportMutation$data;
  variables: AccessReviewCampaignDetailPageExportMutation$variables;
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
    "concreteType": "ExportCampaignEvidencePayload",
    "kind": "LinkedField",
    "name": "exportCampaignEvidence",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "checksumSha256",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "payload",
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
    "name": "AccessReviewCampaignDetailPageExportMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewCampaignDetailPageExportMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "1245fb63f0a3d9bc2f2eee86457084b3",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignDetailPageExportMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewCampaignDetailPageExportMutation(\n  $input: ExportCampaignEvidenceInput!\n) {\n  exportCampaignEvidence(input: $input) {\n    checksumSha256\n    payload\n  }\n}\n"
  }
};
})();

(node as any).hash = "e23760ed87d0b321144a02955e930715";

export default node;
