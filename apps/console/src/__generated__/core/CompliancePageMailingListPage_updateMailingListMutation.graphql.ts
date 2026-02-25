/**
 * @generated SignedSource<<c5e3b54834d5bcf46746de91d32e964a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateMailingListInput = {
  id: string;
  replyTo?: string | null | undefined;
};
export type CompliancePageMailingListPage_updateMailingListMutation$variables = {
  input: UpdateMailingListInput;
};
export type CompliancePageMailingListPage_updateMailingListMutation$data = {
  readonly updateMailingList: {
    readonly mailingList: {
      readonly id: string;
      readonly replyTo: string | null | undefined;
    };
  };
};
export type CompliancePageMailingListPage_updateMailingListMutation = {
  response: CompliancePageMailingListPage_updateMailingListMutation$data;
  variables: CompliancePageMailingListPage_updateMailingListMutation$variables;
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
    "concreteType": "UpdateMailingListPayload",
    "kind": "LinkedField",
    "name": "updateMailingList",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "MailingList",
        "kind": "LinkedField",
        "name": "mailingList",
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
            "name": "replyTo",
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
    "name": "CompliancePageMailingListPage_updateMailingListMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageMailingListPage_updateMailingListMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "30b67540b107297075df95ee044714ea",
    "id": null,
    "metadata": {},
    "name": "CompliancePageMailingListPage_updateMailingListMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageMailingListPage_updateMailingListMutation(\n  $input: UpdateMailingListInput!\n) {\n  updateMailingList(input: $input) {\n    mailingList {\n      id\n      replyTo\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "0a1e6524623210147ba5db2affec456c";

export default node;
