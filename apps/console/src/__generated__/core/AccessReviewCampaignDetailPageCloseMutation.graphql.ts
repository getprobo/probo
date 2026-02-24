/**
 * @generated SignedSource<<7b7e862397b4147f0e61a1e8a5d8c1f8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AccessReviewCampaignStatus = "CANCELLED" | "COMPLETED" | "DRAFT" | "FAILED" | "IN_PROGRESS" | "PENDING_ACTIONS";
export type CloseAccessReviewCampaignInput = {
  accessReviewCampaignId: string;
};
export type AccessReviewCampaignDetailPageCloseMutation$variables = {
  input: CloseAccessReviewCampaignInput;
};
export type AccessReviewCampaignDetailPageCloseMutation$data = {
  readonly closeAccessReviewCampaign: {
    readonly accessReviewCampaign: {
      readonly completedAt: string | null | undefined;
      readonly id: string;
      readonly status: AccessReviewCampaignStatus;
    };
  };
};
export type AccessReviewCampaignDetailPageCloseMutation = {
  response: AccessReviewCampaignDetailPageCloseMutation$data;
  variables: AccessReviewCampaignDetailPageCloseMutation$variables;
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
    "concreteType": "CloseAccessReviewCampaignPayload",
    "kind": "LinkedField",
    "name": "closeAccessReviewCampaign",
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
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "completedAt",
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
    "name": "AccessReviewCampaignDetailPageCloseMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewCampaignDetailPageCloseMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "3cdfd195e50f1e1e12d6110636f6b9a5",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignDetailPageCloseMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewCampaignDetailPageCloseMutation(\n  $input: CloseAccessReviewCampaignInput!\n) {\n  closeAccessReviewCampaign(input: $input) {\n    accessReviewCampaign {\n      id\n      status\n      completedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e3c054bf839928cddbeac593ecc12977";

export default node;
