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

import {
	NodeConnectionTypes,
	type IExecuteFunctions,
	type INodeExecutionData,
	type INodeType,
	type INodeTypeDescription,
} from 'n8n-workflow';
import {
	getAllResourceOperations,
	getAllResourceFields,
	getExecuteFunction,
} from './actions';

export class Probo implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Probo',
		name: 'probo',
		icon: { light: 'file:../../icons/probo-light.svg', dark: 'file:../../icons/probo.svg' },
		group: ['input'],
		version: 1,
		subtitle: '={{$parameter["resource"]}} / {{$parameter["operation"]}}',
		description: 'Consume data from the Probo API',
		defaults: {
			name: 'Probo',
		},
		usableAsTool: true,
		inputs: [NodeConnectionTypes.Main],
		outputs: [NodeConnectionTypes.Main],
		credentials: [
			{
				name: 'proboApi',
				required: true,
				displayOptions: {
					show: {
						authentication: ['apiKey'],
					},
				},
			},
		],
		requestDefaults: {
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json',
			},
		},
		properties: [
			{
				displayName: 'Authentication',
				name: 'authentication',
				type: 'options',
				options: [
					{
						name: 'API Key',
						value: 'apiKey',
					},
				],
				default: 'apiKey',
			},
			{
				displayName: 'Resource',
				name: 'resource',
				type: 'options',
				noDataExpression: true,
				options: [
					{
						name: 'Access Review',
						value: 'accessReview',
						description: 'Manage access review campaigns',
					},
					{
						name: 'Asset',
						value: 'asset',
						description: 'Manage assets',
					},
					{
						name: 'Audit',
						value: 'audit',
						description: 'Manage audits',
					},
					{
						name: 'Audit Log',
						value: 'auditLog',
						description: 'View audit log entries',
					},
					{
						name: 'Control',
						value: 'control',
						description: 'Manage controls',
					},
					{
						name: 'Data',
						value: 'datum',
						description: 'Manage data',
					},
					{
						name: 'Document',
						value: 'document',
						description: 'Manage documents, versions, and signatures',
					},
					{
						name: 'DPIA',
						value: 'dpia',
						description: 'Manage data protection impact assessments',
					},
					{
						name: 'Evidence',
						value: 'evidence',
						description: 'Manage evidences',
					},
					{
						name: 'Execute',
						value: 'execute',
						description: 'Execute a GraphQL query or mutation',
					},
					{
						name: 'Finding',
						value: 'finding',
						description: 'Manage findings',
					},
					{
						name: 'Framework',
						value: 'framework',
						description: 'Manage frameworks',
					},
					{
						name: 'Measure',
						value: 'measure',
						description: 'Manage measures',
					},
					{
						name: 'Obligation',
						value: 'obligation',
						description: 'Manage obligations',
					},
					{
						name: 'Organization',
						value: 'organization',
						description: 'Manage organizations',
					},
					{
						name: 'Organization Context',
						value: 'organizationContext',
						description: 'Manage organization context',
					},
					{
						name: 'Processing Activity',
						value: 'processingActivity',
						description: 'Manage processing activities',
					},
					{
						name: 'Rights Request',
						value: 'rightsRequest',
						description: 'Manage rights requests',
					},
					{
						name: 'Risk',
						value: 'risk',
						description: 'Manage risks',
					},
					{
						name: 'Snapshot',
						value: 'snapshot',
						description: 'Manage snapshots',
					},
					{
						name: 'Statement of Applicability',
						value: 'statementOfApplicability',
						description: 'Manage statements of applicability',
					},
					{
						name: 'Task',
						value: 'task',
						description: 'Manage tasks',
					},
					{
						name: 'TIA',
						value: 'tia',
						description: 'Manage transfer impact assessments',
					},
					{
						name: 'Trust Center',
						value: 'trustCenter',
						description: 'Manage trust center',
					},
					{
						name: 'User',
						value: 'user',
						description: 'Manage organization users (profiles)',
					},
					{
						name: 'ThirdParty',
						value: 'thirdParty',
						description: 'Manage thirdParties',
					},
					{
						name: 'Webhook',
						value: 'webhook',
						description: 'Manage webhook subscriptions',
					},
				],
				default: 'execute',
			},
			...getAllResourceOperations(),
			...getAllResourceFields(),
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const returnData: INodeExecutionData[] = [];

		for (let i = 0; i < items.length; i++) {
			const resource = this.getNodeParameter('resource', i) as string;
			const operation = this.getNodeParameter('operation', i, 'execute') as string;

			const executeFunction = getExecuteFunction(resource, operation);
			const result = await executeFunction.call(this, i);
			returnData.push(result);
		}

		return [returnData];
	}

	methods = {
		listSearch: {},
	};
}
