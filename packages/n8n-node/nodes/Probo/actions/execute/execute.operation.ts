// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequest, proboConnectApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'API',
		name: 'api',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['execute'],
			},
		},
		options: [
			{
				name: 'Console API',
				value: 'console',
				description: 'Call the Console GraphQL API (/api/console/v1/graphql)',
			},
			{
				name: 'Connect API',
				value: 'connect',
				description: 'Call the Connect GraphQL API (/api/connect/v1/graphql)',
			},
		],
		default: 'console',
		description: 'Which Probo API to call',
	},
	{
		displayName: 'Query',
		name: 'query',
		type: 'string',
		typeOptions: {
			rows: 5,
		},
		displayOptions: {
			show: {
				resource: ['execute'],
			},
		},
		default: '',
		description: 'The complete GraphQL operation including operation name and variable declarations (e.g., "query GetUser($userId: ID!) { node(ID: $userId) { ID } }" or "mutation UpdateUser($input: UpdateUserInput!) { updateUser(input: $input) { ID } }")',
		required: true,
	},
	{
		displayName: 'Variables',
		name: 'variables',
		type: 'json',
		displayOptions: {
			show: {
				resource: ['execute'],
			},
		},
		default: '{}',
		description: 'GraphQL variables as JSON object',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const api = this.getNodeParameter('api', itemIndex, 'console') as string;
	const query = this.getNodeParameter('query', itemIndex) as string;
	const variablesParam = this.getNodeParameter('variables', itemIndex) as string;

	// Basic validation: check if query contains a GraphQL operation
	const trimmedQuery = query.trim();
	if (!trimmedQuery) {
		throw new NodeOperationError(this.getNode(), 'GraphQL query cannot be empty', { itemIndex });
	}

	// Check for operation type (query, mutation, or subscription)
	const operationMatch = trimmedQuery.match(/^\s*(query|mutation|subscription)\s+(\w+)/i);
	if (!operationMatch) {
		throw new NodeOperationError(
			this.getNode(),
			'GraphQL operation must start with "query", "mutation", or "subscription" followed by an operation name (e.g., "query GetUser { ... }" or "mutation UpdateUser { ... }")',
			{ itemIndex },
		);
	}

	let variables = {};
	if (variablesParam) {
		try {
			variables =
				typeof variablesParam === 'string' ? JSON.parse(variablesParam) : variablesParam;
		} catch (error) {
			throw new NodeOperationError(this.getNode(), error as Error, {
				itemIndex,
				description: 'Invalid JSON in Variables',
			});
		}
	}

	const requestFn = api === 'connect' ? proboConnectApiRequest : proboApiRequest;
	const responseData = await requestFn.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
