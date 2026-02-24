/**
 * @generated SignedSource<<ec22f1323cacabf5592a7847140ecb4a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AccessReviewCampaignSourceFetchStatus = "FAILED" | "FETCHING" | "QUEUED" | "SUCCESS";
export type AccessReviewCampaignStatus = "CANCELLED" | "COMPLETED" | "DRAFT" | "FAILED" | "IN_PROGRESS" | "PENDING_ACTIONS";
export type RetryStartAccessReviewCampaignInput = {
  accessReviewCampaignId: string;
};
export type AccessReviewCampaignDetailPageRetryStartMutation$variables = {
  input: RetryStartAccessReviewCampaignInput;
};
export type AccessReviewCampaignDetailPageRetryStartMutation$data = {
  readonly retryStartAccessReviewCampaign: {
    readonly accessReviewCampaign: {
      readonly id: string;
      readonly scopeSources: ReadonlyArray<{
        readonly attemptCount: number;
        readonly fetchCompletedAt: string | null | undefined;
        readonly fetchStartedAt: string | null | undefined;
        readonly fetchStatus: AccessReviewCampaignSourceFetchStatus;
        readonly fetchedAccountsCount: number;
        readonly id: string;
        readonly lastError: string | null | undefined;
        readonly name: string;
      }>;
      readonly startedAt: string | null | undefined;
      readonly status: AccessReviewCampaignStatus;
    };
  };
};
export type AccessReviewCampaignDetailPageRetryStartMutation = {
  response: AccessReviewCampaignDetailPageRetryStartMutation$data;
  variables: AccessReviewCampaignDetailPageRetryStartMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "RetryStartAccessReviewCampaignPayload",
    "kind": "LinkedField",
    "name": "retryStartAccessReviewCampaign",
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
          (v1/*: any*/),
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
            "name": "startedAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "AccessReviewCampaignScopeSource",
            "kind": "LinkedField",
            "name": "scopeSources",
            "plural": true,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "name",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fetchStatus",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fetchedAccountsCount",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "attemptCount",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "lastError",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fetchStartedAt",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fetchCompletedAt",
                "storageKey": null
              }
            ],
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
    "name": "AccessReviewCampaignDetailPageRetryStartMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewCampaignDetailPageRetryStartMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "20618e4a7bf686ac9fb12d36322c4897",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignDetailPageRetryStartMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewCampaignDetailPageRetryStartMutation(\n  $input: RetryStartAccessReviewCampaignInput!\n) {\n  retryStartAccessReviewCampaign(input: $input) {\n    accessReviewCampaign {\n      id\n      status\n      startedAt\n      scopeSources {\n        id\n        name\n        fetchStatus\n        fetchedAccountsCount\n        attemptCount\n        lastError\n        fetchStartedAt\n        fetchCompletedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "fd620ffcabaad3bfc06174e7656c21df";

export default node;
