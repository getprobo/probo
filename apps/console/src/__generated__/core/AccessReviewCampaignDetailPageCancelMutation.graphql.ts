/**
 * @generated SignedSource<<456510157785fc9d8cc413ffd5d577e7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AccessReviewCampaignStatus = "CANCELLED" | "COMPLETED" | "DRAFT" | "FAILED" | "IN_PROGRESS" | "PENDING_ACTIONS";
export type CancelAccessReviewCampaignInput = {
  accessReviewCampaignId: string;
};
export type AccessReviewCampaignDetailPageCancelMutation$variables = {
  input: CancelAccessReviewCampaignInput;
};
export type AccessReviewCampaignDetailPageCancelMutation$data = {
  readonly cancelAccessReviewCampaign: {
    readonly accessReviewCampaign: {
      readonly id: string;
      readonly status: AccessReviewCampaignStatus;
    };
  };
};
export type AccessReviewCampaignDetailPageCancelMutation = {
  response: AccessReviewCampaignDetailPageCancelMutation$data;
  variables: AccessReviewCampaignDetailPageCancelMutation$variables;
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
    "concreteType": "CancelAccessReviewCampaignPayload",
    "kind": "LinkedField",
    "name": "cancelAccessReviewCampaign",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "AccessReviewCampaign",
        "kind": "LinkedField",
        "name": "accessReviewCampaign",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "status",
            "storageKey": null
          }
        ],
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
    "name": "AccessReviewCampaignDetailPageCancelMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewCampaignDetailPageCancelMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "d078ec192970bd5ad3026b7816146644",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignDetailPageCancelMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewCampaignDetailPageCancelMutation(\n  $input: CancelAccessReviewCampaignInput!\n) {\n  cancelAccessReviewCampaign(input: $input) {\n    accessReviewCampaign {\n      id\n      status\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "4aacdb01cc7f2f22037abc861f07142e";

export default node;
