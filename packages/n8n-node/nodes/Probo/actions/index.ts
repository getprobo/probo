// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import * as accessReview from './accessReview';
import * as asset from './asset';
import * as audit from './audit';
import * as auditLog from './auditLog';
import * as control from './control';
import * as datum from './datum';
import * as document from './document';
import * as dpia from './dpia';
import * as evidence from './evidence';
import * as execute from './execute';
import * as finding from './finding';
import * as framework from './framework';
import * as measure from './measure';
import * as obligation from './obligation';
import * as organization from './organization';
import * as organizationContext from './organizationContext';
import * as processingActivity from './processingActivity';
import * as rightsRequest from './rightsRequest';
import * as user from './user';
import * as risk from './risk';
import * as snapshot from './snapshot';
import * as statementOfApplicability from './statementOfApplicability';
import * as task from './task';
import * as tia from './tia';
import * as trustCenter from './trustCenter';
import * as vendor from './vendor';
import * as webhook from './webhook';

export interface ResourceModule {
	description: INodeProperties[];
	[key: string]: OperationModule | INodeProperties[];
}

export interface OperationModule {
	description: INodeProperties[];
	execute: (this: IExecuteFunctions, itemIndex: number) => Promise<INodeExecutionData>;
}

export const resources: Record<string, ResourceModule> = {
	accessReview: accessReview as ResourceModule,
	asset: asset as ResourceModule,
	audit: audit as ResourceModule,
	auditLog: auditLog as ResourceModule,
	control: control as ResourceModule,
	datum: datum as ResourceModule,
	document: document as ResourceModule,
	dpia: dpia as ResourceModule,
	evidence: evidence as ResourceModule,
	execute: execute as ResourceModule,
	finding: finding as ResourceModule,
	framework: framework as ResourceModule,
	measure: measure as ResourceModule,
	obligation: obligation as ResourceModule,
	organization: organization as ResourceModule,
	organizationContext: organizationContext as ResourceModule,
	processingActivity: processingActivity as ResourceModule,
	rightsRequest: rightsRequest as ResourceModule,
	user: user as ResourceModule,
	risk: risk as ResourceModule,
	snapshot: snapshot as ResourceModule,
	statementOfApplicability: statementOfApplicability as ResourceModule,
	task: task as ResourceModule,
	tia: tia as ResourceModule,
	trustCenter: trustCenter as ResourceModule,
	vendor: vendor as ResourceModule,
	webhook: webhook as ResourceModule,
};

export function getAllResourceOperations(): INodeProperties[] {
	const operations: INodeProperties[] = [];

	for (const resource of Object.values(resources)) {
		const operationProp = resource.description.find((prop) => prop.name === 'operation');
		if (operationProp) {
			operations.push(operationProp);
		}
	}

	return operations;
}

export function getAllResourceFields(): INodeProperties[] {
	const fields: INodeProperties[] = [];

	for (const resource of Object.values(resources)) {
		fields.push(...resource.description.filter((prop) => prop.name !== 'operation'));
	}

	return fields;
}

export function getExecuteFunction(resourceName: string, operationName: string) {
	const resource = resources[resourceName];
	if (!resource) {
		throw new Error(`Unknown resource: ${resourceName}`);
	}

	const operationKey = resourceName === 'execute' ? 'execute' : operationName;

	const operation = resource[operationKey] as OperationModule;

	if (!operation || typeof operation.execute !== 'function') {
		throw new Error(`Unknown operation: ${operationName} for resource: ${resourceName}`);
	}

	return operation.execute;
}
