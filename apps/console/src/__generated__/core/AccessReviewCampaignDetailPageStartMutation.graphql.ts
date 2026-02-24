/**
 * @generated SignedSource<<7f148bcfc1b46b8888687ecbef11fcb4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AccessReviewCampaignSourceFetchStatus = "FAILED" | "FETCHING" | "QUEUED" | "SUCCESS";
export type AccessReviewCampaignStatus = "CANCELLED" | "COMPLETED" | "DRAFT" | "FAILED" | "IN_PROGRESS" | "PENDING_ACTIONS";
export type StartAccessReviewCampaignInput = {
  accessReviewCampaignId: string;
  accessSourceIds?: ReadonlyArray<string> | null | undefined;
};
export type AccessReviewCampaignDetailPageStartMutation$variables = {
  input: StartAccessReviewCampaignInput;
};
export type AccessReviewCampaignDetailPageStartMutation$data = {
  readonly startAccessReviewCampaign: {
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
export type AccessReviewCampaignDetailPageStartMutation = {
  response: AccessReviewCampaignDetailPageStartMutation$data;
  variables: AccessReviewCampaignDetailPageStartMutation$variables;
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
    "concreteType": "StartAccessReviewCampaignPayload",
    "kind": "LinkedField",
    "name": "startAccessReviewCampaign",
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
    "name": "AccessReviewCampaignDetailPageStartMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewCampaignDetailPageStartMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "08197e8ea7d8c7efbb24352edbd2201e",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignDetailPageStartMutation",
    "operationKind": "mutation",
    "text": "mutation AccessReviewCampaignDetailPageStartMutation(\n  $input: StartAccessReviewCampaignInput!\n) {\n  startAccessReviewCampaign(input: $input) {\n    accessReviewCampaign {\n      id\n      status\n      startedAt\n      scopeSources {\n        id\n        name\n        fetchStatus\n        fetchedAccountsCount\n        attemptCount\n        lastError\n        fetchStartedAt\n        fetchCompletedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6fce7aa7491fcdc0c78712e661924a56";

export default node;
