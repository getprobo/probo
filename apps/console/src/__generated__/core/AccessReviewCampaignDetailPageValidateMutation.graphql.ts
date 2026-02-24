/**
 * @generated SignedSource<<be18bc53817a6d6217130f5f36b9404f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AccessReviewCampaignStatus = "CANCELLED" | "COMPLETED" | "DRAFT" | "FAILED" | "IN_PROGRESS" | "PENDING_ACTIONS";
export type ValidateAccessReviewCampaignInput = {
  accessReviewCampaignId: string;
  note?: string | null | undefined;
};
export type AccessReviewCampaignDetailPageValidateMutation$variables = {
  input: ValidateAccessReviewCampaignInput;
};
export type AccessReviewCampaignDetailPageValidateMutation$data = {
  readonly validateAccessReviewCampaign: {
    readonly accessReviewCampaign: {
      readonly id: string;
      readonly status: AccessReviewCampaignStatus;
    };
  };
};
export type AccessReviewCampaignDetailPageValidateMutation = {
  response: AccessReviewCampaignDetailPageValidateMutation$data;
  variables: AccessReviewCampaignDetailPageValidateMutation$variables;
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
    "concreteType": "ValidateAccessReviewCampaignPayload",
    "kind": "LinkedField",
    "name": "validateAccessReviewCampaign",
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
    "name": "AccessReviewCampaignDetailPageValidateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewCampaignDetailPageValidateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "6ab774604d1ad1e9d8fca3b800b09e6d",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignDetailPageValidateMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewCampaignDetailPageValidateMutation(\n  $input: ValidateAccessReviewCampaignInput!\n) {\n  validateAccessReviewCampaign(input: $input) {\n    accessReviewCampaign {\n      id\n      status\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "01cd1b6320c46fec9b28e11a20f36419";

export default node;
