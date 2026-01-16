/**
 * @generated SignedSource<<be978f367e58cceaa6d2d3d581dadd53>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RightsRequestState = "DONE" | "IN_PROGRESS" | "TODO";
export type RightsRequestType = "ACCESS" | "DELETION" | "PORTABILITY";
export type UpdateRightsRequestInput = {
  actionTaken?: string | null | undefined;
  contact?: string | null | undefined;
  dataSubject?: string | null | undefined;
  deadline?: string | null | undefined;
  details?: string | null | undefined;
  id: string;
  requestState?: RightsRequestState | null | undefined;
  requestType?: RightsRequestType | null | undefined;
};
export type RightsRequestGraphUpdateMutation$variables = {
  input: UpdateRightsRequestInput;
};
export type RightsRequestGraphUpdateMutation$data = {
  readonly updateRightsRequest: {
    readonly rightsRequest: {
      readonly actionTaken: string | null | undefined;
      readonly contact: string | null | undefined;
      readonly dataSubject: string | null | undefined;
      readonly deadline: string | null | undefined;
      readonly details: string | null | undefined;
      readonly id: string;
      readonly requestState: RightsRequestState;
      readonly requestType: RightsRequestType;
      readonly updatedAt: string;
    };
  };
};
export type RightsRequestGraphUpdateMutation = {
  response: RightsRequestGraphUpdateMutation$data;
  variables: RightsRequestGraphUpdateMutation$variables;
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
    "concreteType": "UpdateRightsRequestPayload",
    "kind": "LinkedField",
    "name": "updateRightsRequest",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "RightsRequest",
        "kind": "LinkedField",
        "name": "rightsRequest",
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
            "name": "requestType",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "requestState",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "dataSubject",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "contact",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "details",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "deadline",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "actionTaken",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
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
    "name": "RightsRequestGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "RightsRequestGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "4b06cb49236bb946cfa333e95b80f2a8",
    "id": null,
    "metadata": {},
    "name": "RightsRequestGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation RightsRequestGraphUpdateMutation(\n  $input: UpdateRightsRequestInput!\n) {\n  updateRightsRequest(input: $input) {\n    rightsRequest {\n      id\n      requestType\n      requestState\n      dataSubject\n      contact\n      details\n      deadline\n      actionTaken\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f2f42ebc57de95ec4a760e7d701f90fa";

export default node;
