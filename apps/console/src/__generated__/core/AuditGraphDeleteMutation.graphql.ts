/**
 * @generated SignedSource<<abf5473524b2c61dd92c4975aa4b21f1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteAuditInput = {
  auditId: string;
};
export type AuditGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteAuditInput;
};
export type AuditGraphDeleteMutation$data = {
  readonly deleteAudit: {
    readonly deletedAuditId: string;
  };
};
export type AuditGraphDeleteMutation = {
  response: AuditGraphDeleteMutation$data;
  variables: AuditGraphDeleteMutation$variables;
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
  "name": "deletedAuditId",
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
    "name": "AuditGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteAuditPayload",
        "kind": "LinkedField",
        "name": "deleteAudit",
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
    "name": "AuditGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteAuditPayload",
        "kind": "LinkedField",
        "name": "deleteAudit",
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
            "name": "deletedAuditId",
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
    "cacheID": "48906b2f55360bea9a8fc7f0daea34be",
    "id": null,
    "metadata": {},
    "name": "AuditGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation AuditGraphDeleteMutation(\n  $input: DeleteAuditInput!\n) {\n  deleteAudit(input: $input) {\n    deletedAuditId\n  }\n}\n"
  }
};
})();

(node as any).hash = "e0f3771289e512982adc900199820821";

export default node;
