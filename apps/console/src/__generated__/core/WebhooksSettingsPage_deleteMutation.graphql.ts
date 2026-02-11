/**
 * @generated SignedSource<<86d94ecf595a0048971276541e185905>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteWebhookConfigurationInput = {
  webhookConfigurationId: string;
};
export type WebhooksSettingsPage_deleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteWebhookConfigurationInput;
};
export type WebhooksSettingsPage_deleteMutation$data = {
  readonly deleteWebhookConfiguration: {
    readonly deletedWebhookConfigurationId: string;
  };
};
export type WebhooksSettingsPage_deleteMutation = {
  response: WebhooksSettingsPage_deleteMutation$data;
  variables: WebhooksSettingsPage_deleteMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedWebhookConfigurationId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "WebhooksSettingsPage_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteWebhookConfigurationPayload",
        "kind": "LinkedField",
        "name": "deleteWebhookConfiguration",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "WebhooksSettingsPage_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteWebhookConfigurationPayload",
        "kind": "LinkedField",
        "name": "deleteWebhookConfiguration",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedWebhookConfigurationId",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "ae53fa4e16e4491720c1fdfc52301c09",
    "id": null,
    "metadata": {},
    "name": "WebhooksSettingsPage_deleteMutation",
    "operationKind": "mutation",
    "text": "mutation WebhooksSettingsPage_deleteMutation(\n  $input: DeleteWebhookConfigurationInput!\n) {\n  deleteWebhookConfiguration(input: $input) {\n    deletedWebhookConfigurationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "632fce2d89360bed2d857dd3feb90b9d";

export default node;
